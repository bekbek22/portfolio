-- name: ListTagsWithCounts :many
-- Tags used by at least one published project, by category (enum order) then name.
SELECT tg.slug, tg.name, tg.category, count(*)::int AS project_count
FROM tags tg
JOIN project_tags pt ON pt.tag_id = tg.id
JOIN projects p ON p.id = pt.project_id AND p.published
GROUP BY tg.id
ORDER BY tg.category, tg.name;

-- name: UpsertTag :one
INSERT INTO tags (slug, name, category)
VALUES (@slug, @name, @category)
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name, category = EXCLUDED.category
RETURNING id;
