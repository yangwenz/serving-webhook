CREATE TABLE "task"
(
    "id"           bigserial PRIMARY KEY,
    "task_id"      varchar UNIQUE NOT NULL,
    "user_id"      varchar,
    "model_name"   varchar        NOT NULL,
    "created_at"   timestamptz    NOT NULL DEFAULT (now()),
    "updated_at"   timestamptz    NOT NULL DEFAULT (now()),
    "running_time" float,
    "status"       varchar
);