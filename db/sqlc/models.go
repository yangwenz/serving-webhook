// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Task struct {
	ID          int64         `json:"id"`
	TaskID      string        `json:"task_id"`
	UserID      pgtype.Text   `json:"user_id"`
	ModelName   string        `json:"model_name"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	RunningTime pgtype.Float8 `json:"running_time"`
	Status      pgtype.Text   `json:"status"`
}