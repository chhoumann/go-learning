-- name: CreatePost :one
INSERT INTO posts (title, url, description, published_at, feed_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at, updated_at, title, url, description, published_at, feed_id;

-- name: GetPostsByUser :many
SELECT p.id, p.created_at, p.updated_at, p.title, p.url, p.description, p.published_at, p.feed_id
FROM posts p
JOIN feeds f ON p.feed_id = f.id
JOIN users u ON f.user_id = u.id
WHERE u.id = $1
ORDER BY p.published_at DESC
LIMIT $2;
