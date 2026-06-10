-- name: CreatePost :exec
INSERT INTO posts(id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT(url) DO NOTHING;

-- name: GetPostForUser :many
SELECT p.*, f.name as feed_name
FROM posts p
JOIN feed_follows ff ON ff.feed_id = p.feed_id
JOIN feeds f ON p.feed_id = f.id
WHERE ff.user_id = $1
ORDER BY p.published_at DESC;

-- name: MarkPostAsRead :exec
UPDATE posts SET read = TRUE, UPDATED_AT = NOW() WHERE id = $1;

-- name: MarkPostAsUnread :exec
UPDATE posts SET read = FALSE, UPDATED_AT = NOW() WHERE id = $1;

-- name: GetUnreadNumberForUser :one
SELECT COUNT(*) FROM posts p
JOIN feed_follows ff ON ff.feed_id = p.feed_id
WHERE ff.user_id = $1 AND p.read = False;