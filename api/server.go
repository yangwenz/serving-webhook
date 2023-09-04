package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yangwenz/model-webhook/storage"
	"github.com/yangwenz/model-webhook/utils"
	"net/http"
)

type Server struct {
	config   utils.Config
	router   *gin.Engine
	uploader storage.Store
}

func NewServer(config utils.Config, uploader storage.Store) (*Server, error) {
	server := Server{
		config:   config,
		router:   nil,
		uploader: uploader,
	}
	server.setupRouter()
	return &server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	router.MaxMultipartMemory = 16 << 20 // 16 MiB

	router.GET("/live", server.checkHealth)
	router.GET("/ready", server.checkHealth)
	router.POST("/upload", server.Upload)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func (server *Server) checkHealth(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "API OK"})
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
