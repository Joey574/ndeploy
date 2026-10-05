-- name: CreateNode :one
INSERT INTO nodes (user, host, host_key, identity_file)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: DeleteNode :exec
DELETE FROM nodes WHERE id = ?;

-- name: GetNode :one
SELECT * FROM nodes WHERE id = ?;

-- name: GetNodeByHost :one
SELECT * FROM nodes WHERE host = ?;

-- name: ListNodes :many
SELECT * FROM nodes ORDER BY id;

-- name: CountNodes :one
SELECT COUNT(DISTINCT host) FROM nodes;

-- name: UpdateNodeUser :exec
UPDATE nodes SET user = ? WHERE id = ?;

-- name: UpdateNodeHostKey :exec
UPDATE nodes SET host_key = ? WHERE id = ?;

-- name: UpdateNodeIdentityFile :exec
UPDATE nodes SET identity_file = ? WHERE id = ?;

-- name: UpdateNodeConfigFile :exec
UPDATE nodes SET config_file = ? WHERE id = ?;

-- name: UpdateNodeHost :exec
UPDATE nodes SET host = ? WHERE id = ?;

-- name: UpdateNode :exec
UPDATE nodes
SET user = ?, host = ?, host_key = ?, identity_file = ?, config_file = ?
WHERE id = ?;
