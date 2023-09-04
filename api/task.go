package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type TaskInfo struct {
	ID           string      `json:"id"`
	ModelName    string      `json:"model_name"`
	ModelVersion string      `json:"model_version"`
	Status       string      `json:"status"`
	Runtime      string      `json:"runtime"`
	CreatedAt    time.Time   `json:"created_at"`
	Outputs      interface{} `json:"outputs"`
}

type CreateRequest struct {
	ModelName    string `json:"model_name"`
	ModelVersion string `json:"model_version"`
}

type UpdateRequest struct {
	ID      string      `json:"id" binding:"required"`
	Status  string      `json:"status"`
	Runtime string      `json:"runtime"`
	Outputs interface{} `json:"outputs"`
}

type URI struct {
	ID string `json:"id" uri:"id"`
}

func (server *Server) KeyDuration() time.Duration {
	duration, _ := time.ParseDuration("48h")
	return duration
}

func (server *Server) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	task := &TaskInfo{
		ID:           uuid.New().String(),
		ModelName:    req.ModelName,
		ModelVersion: req.ModelVersion,
		Status:       "pending",
		Runtime:      "",
		CreatedAt:    time.Now(),
		Outputs:      nil,
	}
	duration := server.KeyDuration()
	err := server.cache.SetKey(task.ID, task, duration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
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
	value, err := server.cache.GetKey(req.ID)
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
	if req.Status != "" {
		task.Status = req.Status
	}
	if req.Runtime != "" {
		task.Runtime = req.Runtime
	}
	if req.Outputs != nil {
		task.Outputs = req.Outputs
	}
	duration := server.KeyDuration()
	err = server.cache.SetKey(task.ID, task, duration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": task.ID})
}
