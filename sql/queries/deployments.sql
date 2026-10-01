-- name: CreateDeployment :one
INSERT INTO deployments (target_host, build_host, is_upgrade)
VALUES (?, ?, ?)
RETURNING *;

-- name: SetCancelled :exec
UPDATE nodes SET is_cancelled = ? WHERE id = ?;
