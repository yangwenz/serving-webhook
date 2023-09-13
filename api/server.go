package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yangwenz/model-webhook/storage"
	"github.com/yangwenz/model-webhook/utils"
	"net/http"
)

type Server struct {
	config utils.Config
	router *gin.Engine
	store  storage.Store
	cache  storage.Cache
}

func NewServer(config utils.Config, store storage.Store, cache storage.Cache) (*Server, error) {
	server := Server{
		config: config,
		router: nil,
		store:  store,
		cache:  cache,
	}
	server.setupRouter()
	return &server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	router.MaxMultipartMemory = 16 << 20 // 16 MiB

	router.GET("/live", server.checkHealth)
	router.GET("/ready", server.checkHealth)

	taskRoutes := router.Group("/").Use(authMiddleware(server.config))
	taskRoutes.POST("/upload", server.Upload)
	taskRoutes.POST("/task", server.Create)
	taskRoutes.GET("/task/:id", server.Get)
	taskRoutes.PUT("/task", server.Update)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func (server *Server) checkHealth(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "API OK"})
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
