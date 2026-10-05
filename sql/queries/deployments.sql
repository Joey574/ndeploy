-- name: CreateDeployment :one
INSERT INTO deployments (target_node, build_node, action, config_file, command, is_upgrade, is_rollback)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: FinishDeployment :exec
UPDATE deployments
SET exit_code = ?, error = ?, finished_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: SetCancelled :exec
UPDATE deployments SET is_cancelled = ? WHERE id = ?;

-- name: GetDeployment :one
SELECT * FROM deployments WHERE id = ?;

-- name: ListDeployments :many
SELECT * FROM deployments
ORDER BY executed_at DESC, id DESC
LIMIT ?;

-- name: ListDeploymentsForNode :many
SELECT * FROM deployments
WHERE target_node = ?
ORDER BY executed_at DESC, id DESC
LIMIT ?;
