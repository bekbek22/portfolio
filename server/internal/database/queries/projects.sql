-- name: ListPublishedProjects :many
-- Published projects that have ALL of @tag_slugs (none = no tag filter), optionally
-- filtered by featured and a search term. @search must already be LIKE-escaped (\ % _).
SELECT p.id, p.slug, p.title, p.summary, p.cover_image_url, p.featured,
       COALESCE(t.tags, '[]'::json) AS tags
FROM projects p
LEFT JOIN LATERAL (
    SELECT json_agg(json_build_object('slug', tg.slug, 'name', tg.name) ORDER BY tg.name) AS tags
    FROM project_tags pt
    JOIN tags tg ON tg.id = pt.tag_id
    WHERE pt.project_id = p.id
) t ON true
WHERE p.published
  AND (sqlc.narg('featured')::boolean IS NULL OR p.featured = sqlc.narg('featured'))
  AND (sqlc.narg('search')::text IS NULL
       OR p.title   ILIKE '%' || sqlc.narg('search') || '%' ESCAPE '\'
       OR p.summary ILIKE '%' || sqlc.narg('search') || '%' ESCAPE '\')
  AND (COALESCE(cardinality(@tag_slugs::text[]), 0) = 0 OR p.id IN (
        -- Intersection: keep projects matching every requested tag.
        SELECT pt.project_id
        FROM project_tags pt
        JOIN tags tg ON tg.id = pt.tag_id
        WHERE tg.slug = ANY(@tag_slugs::text[])
        GROUP BY pt.project_id
        HAVING count(DISTINCT tg.slug) = cardinality(@tag_slugs::text[])))
ORDER BY p.sort_order, p.created_at DESC;

-- name: GetPublishedProjectBySlug :one
SELECT p.id, p.slug, p.title, p.summary, p.body, p.cover_image_url, p.repo_url, p.live_url,
       p.featured, p.updated_at,
       COALESCE(t.tags, '[]'::json) AS tags
FROM projects p
LEFT JOIN LATERAL (
    SELECT json_agg(json_build_object('slug', tg.slug, 'name', tg.name) ORDER BY tg.name) AS tags
    FROM project_tags pt
    JOIN tags tg ON tg.id = pt.tag_id
    WHERE pt.project_id = p.id
) t ON true
WHERE p.slug = @slug AND p.published;

-- name: UpsertProject :one
-- Insert or update by slug. updated_at only moves when the content actually changes,
-- so re-running the seed is a no-op.
INSERT INTO projects (slug, title, summary, body, cover_image_url, repo_url, live_url,
                      featured, published, sort_order)
VALUES (@slug, @title, @summary, @body, @cover_image_url, @repo_url, @live_url,
        @featured, @published, @sort_order)
ON CONFLICT (slug) DO UPDATE SET
    title           = EXCLUDED.title,
    summary         = EXCLUDED.summary,
    body            = EXCLUDED.body,
    cover_image_url = EXCLUDED.cover_image_url,
    repo_url        = EXCLUDED.repo_url,
    live_url        = EXCLUDED.live_url,
    featured        = EXCLUDED.featured,
    published       = EXCLUDED.published,
    sort_order      = EXCLUDED.sort_order,
    updated_at      = CASE
        WHEN (projects.title, projects.summary, projects.body, projects.cover_image_url,
              projects.repo_url, projects.live_url, projects.featured, projects.published,
              projects.sort_order)
             IS DISTINCT FROM
             (EXCLUDED.title, EXCLUDED.summary, EXCLUDED.body, EXCLUDED.cover_image_url,
              EXCLUDED.repo_url, EXCLUDED.live_url, EXCLUDED.featured, EXCLUDED.published,
              EXCLUDED.sort_order)
        THEN now()
        ELSE projects.updated_at
    END
RETURNING id;

-- name: DeleteProjectTags :exec
DELETE FROM project_tags WHERE project_id = @project_id;

-- name: AddProjectTags :exec
INSERT INTO project_tags (project_id, tag_id)
SELECT @project_id, unnest(@tag_ids::uuid[]);
