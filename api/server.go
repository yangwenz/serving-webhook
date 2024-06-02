package api

import (
	db "github.com/HyperGAI/serving-webhook/db/sqlc"
	"github.com/HyperGAI/serving-webhook/storage"
	"github.com/HyperGAI/serving-webhook/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Server struct {
	config   utils.Config
	router   *gin.Engine
	store    storage.Store
	cache    storage.Cache
	database db.Store
}

func NewServer(
	config utils.Config,
	store storage.Store,
	cache storage.Cache,
	database db.Store,
) (*Server, error) {
	server := Server{
		config:   config,
		router:   nil,
		store:    store,
		cache:    cache,
		database: database,
	}
	server.setupRouter()
	return &server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	router.MaxMultipartMemory = 32 << 20 // 32 MiB

	router.GET("/live", server.checkHealth)
	router.GET("/ready", server.checkHealth)

	taskRoutes := router.Group("/").Use(authMiddleware(server.config))
	if server.store != nil {
		taskRoutes.POST("/upload", server.Upload)
		taskRoutes.POST("/upload_batch", server.UploadBatch)
	}
	if server.cache != nil {
		taskRoutes.POST("/task", server.Create)
		taskRoutes.GET("/task/:id", server.Get)
		taskRoutes.PUT("/task", server.Update)
		// If database is not set, it will return an empty list
		taskRoutes.GET("/task/modelstatus", server.GetTaskByModelStatus)
	}
	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func (server *Server) Handler() http.Handler {
	return server.router.Handler()
}

func (server *Server) checkHealth(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "API OK"})
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
