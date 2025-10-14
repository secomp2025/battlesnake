-- name: GetCode :one
SELECT *
FROM codes
WHERE id = ?
LIMIT 1;
-- name: FindCode :one
SELECT *
FROM codes
WHERE code = ?
LIMIT 1;
-- name: ListCodes :many
SELECT *
FROM codes
ORDER BY created_at ASC;
-- name: CreateCode :one
INSERT INTO codes (code)
VALUES (?)
RETURNING *;
-- name: UpdateCode :exec
UPDATE codes
SET code = ?
WHERE code = ?
RETURNING *;
-- name: DeleteCode :exec
DELETE FROM codes
WHERE code = ?;
-------- TEAM --------
-- name: GetTeam :one
SELECT *
FROM teams
WHERE id = ?
LIMIT 1;
-- name: GetTeamByCode :one
SELECT *
FROM teams
WHERE code_id = ?
LIMIT 1;
-- name: ListTeams :many
SELECT *
FROM teams
ORDER BY created_at ASC;
-- name: CreateTeam :one
INSERT INTO teams (name, code_id)
VALUES (?, ?)
RETURNING *;
-- name: UpdateTeam :exec
UPDATE teams
SET name = ?,
    code_id = ?
WHERE id = ?
RETURNING *;
-- name: DeleteTeam :exec
DELETE FROM teams
WHERE id = ?;
-------- SNAKE --------
-- name: GetSnake :one
SELECT *
FROM snakes
WHERE id = ?
LIMIT 1;
-- name: ListSnakes :many
SELECT *
FROM snakes
ORDER BY created_at ASC;
-- name: CreateSnake :one
INSERT INTO snakes (path, lang, team_id)
VALUES (?, ?, ?)
RETURNING *;
-- name: UpdateSnake :exec
UPDATE snakes
SET path = ?,
    lang = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;
-- name: DeleteSnake :exec
DELETE FROM snakes
WHERE id = ?;
-- name: ListTeamSnakes :many
SELECT *
FROM snakes
WHERE team_id = ?
ORDER BY updated_at ASC;