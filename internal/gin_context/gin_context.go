package gincontext

import (
	"JudgeCore/internal/global"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func Judge(c *gin.Context) {
	file, err := c.FormFile("code")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}
	if err := c.SaveUploadedFile(file, filepath.Join(global.CodeDir, file.Filename)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "File received"})
}
