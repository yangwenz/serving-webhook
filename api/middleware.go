package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yangwenz/model-webhook/utils"
)

const (
	authorizationHeaderKey = "apikey"
)

func authMiddleware(config utils.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if config.SecretAPIKey == "" {
			ctx.Next()
			return
		}
	}
}
