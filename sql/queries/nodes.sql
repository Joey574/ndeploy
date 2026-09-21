-- name: CreateNode :one
INSERT INTO nodes (user, host, host_key, identity_file)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetNode :one
SELECT * FROM nodes WHERE id = ?;

-- name: ListNodes :many
SELECT * FROM nodes ORDER BY id;

-- name: CountNodes :one
SELECT COUNT(DISTINCT host) FROM nodes;

-- name: UpdateNodeHostKey :exec
UPDATE nodes SET host_key = ? WHERE id = ?;

-- name: UpdateNodeIdentityFile :exec
UPDATE nodes SET identity_file = ? WHERE id = ?;

-- name: GetNodeByHost :one
SELECT * FROM nodes WHERE host = ?;
