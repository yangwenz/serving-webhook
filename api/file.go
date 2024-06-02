package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"sync"
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

	id := uuid.New()
	ext := filepath.Ext(file.Filename)
	// location, err := server.store.Upload(src, id.String()+ext)
	location, err := server.store.PutObject(src, id.String()+ext)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"url": location})
}

type UploadResult struct {
	Index     int
	Location  string
	Error     error
	ErrorCode int
}

func (server *Server) UploadBatch(ctx *gin.Context) {
	s := ctx.Request.Header.Get("NUM_FILES")
	numFiles, e := strconv.Atoi(s)
	if e != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(e))
		return
	}

	files := make([]*multipart.FileHeader, 0)
	for i := 0; i < numFiles; i++ {
		key := fmt.Sprintf("file_%d", i)
		file, err := ctx.FormFile(key)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		files = append(files, file)
	}

	var wg sync.WaitGroup
	resultChannel := make(chan UploadResult, numFiles)
	defer close(resultChannel)

	for index, file := range files {
		wg.Add(1)
		go func(f *multipart.FileHeader, i int) {
			defer wg.Done()
			resultChannel <- server.uploadSingleFile(f, i)
		}(file, index)
	}
	wg.Wait()

	file2location := make(map[int]string)
	for i := 0; i < numFiles; i++ {
		r := <-resultChannel
		if r.Error != nil {
			ctx.JSON(r.ErrorCode, errorResponse(r.Error))
			return
		}
		file2location[r.Index] = r.Location
	}
	locations := make([]string, 0)
	for i := 0; i < numFiles; i++ {
		locations = append(locations, file2location[i])
	}
	ctx.JSON(http.StatusOK, gin.H{"urls": locations})
}

func (server *Server) uploadSingleFile(file *multipart.FileHeader, index int) UploadResult {
	src, err := file.Open()
	if err != nil {
		return UploadResult{
			Index:     index,
			Location:  "",
			Error:     err,
			ErrorCode: http.StatusBadRequest,
		}
	}
	defer src.Close()

	id := uuid.New()
	ext := filepath.Ext(file.Filename)
	// location, err := server.store.Upload(src, id.String()+ext)
	location, err := server.store.PutObject(src, id.String()+ext)
	if err != nil {
		return UploadResult{
			Index:     index,
			Location:  "",
			Error:     err,
			ErrorCode: http.StatusInternalServerError,
		}
	}
	return UploadResult{
		Index:     index,
		Location:  location,
		Error:     nil,
		ErrorCode: 200,
	}
}
