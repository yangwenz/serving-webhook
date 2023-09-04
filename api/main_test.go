package api

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/yangwenz/model-webhook/storage"
	"github.com/yangwenz/model-webhook/utils"
	"os"
	"testing"
)

func newTestServer(t *testing.T, store storage.Store, cache storage.Cache) *Server {
	config := utils.Config{}
	server, err := NewServer(config, store, cache)
	require.NoError(t, err)
	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
