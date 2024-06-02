// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Querier interface {
	CreateTask(ctx context.Context, arg CreateTaskParams) (Task, error)
	DeleteTask(ctx context.Context, taskID string) error
	DeleteTaskBeforeDate(ctx context.Context, createdAt time.Time) error
	GetTaskById(ctx context.Context, taskID string) (Task, error)
	GetTaskByUser(ctx context.Context, userID pgtype.Text) ([]Task, error)
	GetTasksByModelNameAndStatus(ctx context.Context, arg GetTasksByModelNameAndStatusParams) ([]Task, error)
	UpdateTask(ctx context.Context, arg UpdateTaskParams) (Task, error)
}

var _ Querier = (*Queries)(nil)