-- name: GetPublishedProjectBySlug :one
SELECT id, slug, title, summary, body, cover_image_url, repo_url, live_url, featured, updated_at
FROM projects
WHERE slug = @slug AND published;
