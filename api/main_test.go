package api

import (
	db "github.com/HyperGAI/serving-webhook/db/sqlc"
	"github.com/HyperGAI/serving-webhook/storage"
	"github.com/HyperGAI/serving-webhook/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func newTestServer(t *testing.T, store storage.Store, cache storage.Cache, database db.Store) *Server {
	config := utils.Config{}
	server, err := NewServer(config, store, cache, database)
	require.NoError(t, err)
	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
