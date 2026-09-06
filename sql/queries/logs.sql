-- name: CreateUnnamedLog :one
INSERT INTO logs(level, message)
VALUES (?, ?)
RETURNING *;

-- name: CreateNamedLog :one
INSERT INTO logs(level, logger_name, message)
VALUES(?, ?, ?)
RETURNING *;

-- name: CreateUnnamedBlobLog :one
INSERT INTO logs(level, message, metadata)
VALUES (?, ?, ?)
RETURNING *;

-- name: CreateNamedBlobLog :one
INSERT INTO logs(level, logger_name, message, metadata)
VALUES(?, ?, ?, ?)
RETURNING *;
