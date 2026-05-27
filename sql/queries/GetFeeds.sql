-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetFeedByUrl :one
Select * FROM feeds
WHERE url = $1;