-- name: CreateTask :one
INSERT INTO "task" (task_id,
                    user_id,
                    model_name,
                    running_time,
                    status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTaskById :one
SELECT *
FROM "task"
WHERE task_id = $1
LIMIT 1;

-- name: GetTaskByUser :many
SELECT *
FROM "task"
WHERE user_id = $1;

-- name: GetTasksByModelNameAndStatus :many
SELECT *
FROM "task"
WHERE model_name = $1
  AND status = $2;

-- name: UpdateTask :one
UPDATE "task"
SET running_time = COALESCE(sqlc.narg(running_time), running_time),
    status       = COALESCE(sqlc.narg(status), status),
    updated_at   = COALESCE(sqlc.narg(updated_at), updated_at)
WHERE task_id = sqlc.arg(task_id)
RETURNING *;

-- name: DeleteTask :exec
DELETE
FROM "task"
WHERE task_id = $1;

-- name: DeleteTaskBeforeDate :exec
DELETE
FROM "task"
WHERE created_at < $1;
