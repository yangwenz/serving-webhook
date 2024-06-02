package api

import (
	"encoding/json"
	"errors"
	db "github.com/HyperGAI/serving-webhook/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type TaskInfo struct {
	ID           string      `json:"id"`
	ModelName    string      `json:"model_name"`
	ModelVersion string      `json:"model_version"`
	Status       string      `json:"status"`
	RunningTime  string      `json:"running_time"`
	CreatedAt    time.Time   `json:"created_at"`
	Outputs      interface{} `json:"outputs"`
	ErrorInfo    string      `json:"error_info"`
	QueueNum     int         `json:"queue_num"`
	QueueID      string      `json:"queue_id"`
}

type CreateRequest struct {
	ID           string `json:"id" binding:"required"`
	ModelName    string `json:"model_name" binding:"required"`
	ModelVersion string `json:"model_version"`
	QueueNum     int    `json:"queue_num"`
	Status       string `json:"status"`
}

type UpdateRequest struct {
	ID           string      `json:"id" binding:"required"`
	Status       string      `json:"status"`
	RunningTime  string      `json:"running_time"`
	Outputs      interface{} `json:"outputs"`
	ErrorInfo    string      `json:"error_info"`
	QueueID      string      `json:"queue_id"`
	DatabaseOnly bool        `json:"database_only"`
}

type URI struct {
	ID string `json:"id" uri:"id"`
}

type GetTaskFromDBRequest struct {
	ModelName string `json:"model_name" binding:"required"`
	Status    string `json:"status" binding:"required"`
}

func (server *Server) KeyDuration() time.Duration {
	duration, _ := time.ParseDuration(server.config.RedisKeyDuration)
	return duration
}

func (server *Server) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	status := "pending"
	if req.Status != "" {
		status = req.Status
	}
	task := &TaskInfo{
		ID:           req.ID,
		ModelName:    req.ModelName,
		ModelVersion: req.ModelVersion,
		Status:       status,
		RunningTime:  "",
		CreatedAt:    time.Now(),
		Outputs:      nil,
		ErrorInfo:    "",
		QueueNum:     req.QueueNum,
		QueueID:      "",
	}
	duration := server.KeyDuration()
	if server.database != nil {
		// Create a task record in both database and redis
		userID := ctx.Request.Header.Get("UID")
		err := server.database.ExecTx(ctx, func(q *db.Queries) error {
			_, e := q.CreateTask(ctx, db.CreateTaskParams{
				TaskID:      task.ID,
				UserID:      pgtype.Text{String: userID, Valid: true},
				ModelName:   task.ModelName,
				RunningTime: pgtype.Float8{Float64: 0, Valid: true},
				Status:      pgtype.Text{String: task.Status, Valid: true},
			})
			if e != nil {
				return e
			}
			if e := server.cache.SetKey(task.ID, task, duration); e != nil {
				return e
			}
			return nil
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	} else {
		// Create a task record in redis
		err := server.cache.SetKey(task.ID, task, duration)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"id": task.ID})
}

func (server *Server) Get(ctx *gin.Context) {
	var id URI
	if err := ctx.BindUri(&id); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	value, err := server.cache.GetKey(id.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var task TaskInfo
	err = json.Unmarshal([]byte(value), &task)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, task)
}

func (server *Server) Update(ctx *gin.Context) {
	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var task TaskInfo
	if !req.DatabaseOnly {
		value, err := server.cache.GetKey(req.ID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		err = json.Unmarshal([]byte(value), &task)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	} else {
		task.ID = req.ID
	}

	if req.Status != "" {
		task.Status = req.Status
	}
	if req.RunningTime != "" {
		task.RunningTime = req.RunningTime
	}
	if req.Outputs != nil {
		task.Outputs = req.Outputs
	}
	if req.ErrorInfo != "" {
		task.ErrorInfo = req.ErrorInfo
	}
	if req.QueueID != "" {
		task.QueueID = req.QueueID
	}
	duration := server.KeyDuration()

	if server.database != nil {
		// Update the task record in both database and redis
		var runningTime float64 = 0
		if req.RunningTime != "" {
			f := strings.Replace(req.RunningTime, "s", "", -1)
			if s, err := strconv.ParseFloat(f, 64); err == nil {
				runningTime = s
			} else {
				log.Error().Msgf("cannot convert %s to float64", f)
			}
		}
		err := server.database.ExecTx(ctx, func(q *db.Queries) error {
			_, e := q.UpdateTask(ctx, db.UpdateTaskParams{
				RunningTime: pgtype.Float8{Float64: runningTime, Valid: req.RunningTime != ""},
				Status:      pgtype.Text{String: task.Status, Valid: req.Status != ""},
				UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
				TaskID:      task.ID,
			})
			if e != nil {
				return e
			}
			if !req.DatabaseOnly {
				if e := server.cache.SetKey(task.ID, task, duration); e != nil {
					return e
				}
			}
			return nil
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	} else {
		// Update the task record in redis
		err := server.cache.SetKey(task.ID, task, duration)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
	}
	ctx.JSON(http.StatusOK, task)
}

func (server *Server) GetTaskByModelStatus(ctx *gin.Context) {
	if server.database != nil {
		var req GetTaskFromDBRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		tasks, err := server.database.GetTasksByModelNameAndStatus(ctx, db.GetTasksByModelNameAndStatusParams{
			ModelName: req.ModelName,
			Status:    pgtype.Text{String: req.Status, Valid: true},
		})
		if hasError(ctx, err) {
			log.Error().Msgf("failed to find records with model=%s and status=%s", req.ModelName, req.Status)
			return
		}
		ctx.JSON(http.StatusOK, tasks)
	} else {
		tasks := make([]db.Task, 0)
		ctx.JSON(http.StatusOK, tasks)
	}
}

func hasError(ctx *gin.Context, err error) bool {
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return true
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return true
	}
	return false
}
