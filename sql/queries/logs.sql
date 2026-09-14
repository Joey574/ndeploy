-- name: CreateLog :one
INSERT INTO logs(source, level, message, metadata)
VALUES(?, ?, ?, ?)
RETURNING *;

-- name: GetNLogs :many
SELECT * FROM logs
LIMIT ? OFFSET ?;

-- name: GetNLogsFromSource :many
SELECT * FROM logs
WHERE source = ?
LIMIT ? OFFSET ?;
