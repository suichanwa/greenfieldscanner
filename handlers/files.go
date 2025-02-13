// handlers/files.go
package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"cloud-storage/models"

	"github.com/gin-gonic/gin"
)

func (a *App) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}
	defer file.Close()

	userID := c.MustGet("userID").(uint)

	// Create user directory if it doesn't exist
	userDir := filepath.Join("storage", fmt.Sprintf("%d", userID))
	if err := os.MkdirAll(userDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	// Save the file
	filePath := filepath.Join(userDir, header.Filename)
	if err := c.SaveUploadedFile(header, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Calculate file hash
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate file hash"})
		return
	}
	fileHash := hex.EncodeToString(hash.Sum(nil))

	// Save file metadata to database
	fileRecord := models.File{
		UserID:       userID,
		Name:         header.Filename,
		Path:         filePath,
		Size:         header.Size,
		Hash:         fileHash,
		LastModified: time.Now(),
	}
	if err := a.DB.Create(&fileRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file metadata"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}
