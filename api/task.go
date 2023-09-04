package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type TaskInfo struct {
	ID           string      `json:"id" binding:"required"`
	ModelName    string      `json:"model_name"`
	ModelVersion string      `json:"model_version"`
	Status       string      `json:"status" default:"pending"`
	Runtime      string      `json:"runtime"`
	CreatedAt    time.Time   `json:"created_at"`
	Outputs      interface{} `json:"outputs"`
}

type CreateRequest struct {
	ModelName    string `json:"model_name"`
	ModelVersion string `json:"model_version"`
}

type UpdateRequest struct {
	Status  string      `json:"status"`
	Runtime string      `json:"runtime"`
	Outputs interface{} `json:"outputs"`
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
		CreatedAt:    time.Now(),
	}
	duration, _ := time.ParseDuration("48h")
	err := server.cache.SetKey(task.ID, task, duration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": task.ID})
}

func (server *Server) Get(ctx *gin.Context) {

}

func (server *Server) Update(ctx *gin.Context) {

}
