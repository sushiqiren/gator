-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at, updated_at, name, url, user_id;

-- name: GetFeedsWithUserNames :many
SELECT feeds.id, feeds.created_at, feeds.updated_at, feeds.name AS feed_name, feeds.url, users.name AS user_name
FROM feeds
JOIN users ON feeds.user_id = users.id;