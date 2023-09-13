package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/yangwenz/model-webhook/utils"
	"net/http"
	"strings"
)

const (
	authorizationHeaderKey = "apikey"
)

func authMiddleware(config utils.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if config.SecretAPIKey != "" {
			authorizationHeader := ctx.GetHeader(authorizationHeaderKey)

			if len(authorizationHeader) == 0 {
				err := errors.New("authorization header is not provided")
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
				return
			}
			fields := strings.Fields(authorizationHeader)
			if len(fields) != 1 {
				err := errors.New("invalid authorization header format")
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
				return
			}

			accessToken := fields[0]
			if accessToken != config.SecretAPIKey {
				err := errors.New("no permission")
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
				return
			}
		}
		ctx.Next()
	}
}
