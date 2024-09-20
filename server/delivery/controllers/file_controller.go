package controllers

import (
	"cinema-server/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FileController struct{}

func NewFileController() *FileController {
	return &FileController{}
}

func (c *FileController) UploadFiles(ctx *gin.Context) {
	file, _ := ctx.Get("file")
	ctx.JSON(http.StatusOK, file)
}

func (c *FileController) DeleteFile(ctx *gin.Context) {
	var file map[string]string
	if err := ctx.ShouldBindJSON(&file); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := utils.DeleteFile(file["url"]); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, file)
}
