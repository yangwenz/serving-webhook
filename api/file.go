package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

/*
curl -X POST http://localhost:12000/upload \
  -F "file=@/Users/abc/test.zip" \
  -H "Content-Type: multipart/form-data"
*/

func (server *Server) Upload(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	defer src.Close()

	location, err := server.uploader.Upload(src, file.Filename)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"path": location})
}
