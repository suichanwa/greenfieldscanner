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

	// Calculate file hash before saving
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate hash"})
		return
	}
	fileHash := hex.EncodeToString(hash.Sum(nil))

	// Reset file pointer
	file.Seek(0, 0)

	// Check if file already exists
	var existingFile models.File
	if err := a.DB.Where("user_id = ? AND hash = ?", userID, fileHash).First(&existingFile).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "File already exists",
			"file":    existingFile,
		})
		return
	}

	// Save file with hash as filename
	filename := fmt.Sprintf("%s%s", fileHash, filepath.Ext(header.Filename))
	filePath := filepath.Join(userDir, filename)

	outFile, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Save file metadata
	fileRecord := models.File{
		UserID:       userID,
		Name:         header.Filename,
		Path:         filePath,
		Size:         header.Size,
		Hash:         fileHash,
		LastModified: time.Now(),
	}

	if err := a.DB.Create(&fileRecord).Error; err != nil {
		os.Remove(filePath) // Cleanup on DB error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file metadata"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"file":    fileRecord,
	})
}

func (a *App) ListFiles(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var files []models.File
	if err := a.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

func (a *App) DownloadFile(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	fileID := c.Param("id")

	var file models.File
	if err := a.DB.Where("id = ? AND user_id = ?", fileID, userID).First(&file).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.File(file.Path)
}

func (a *App) DeleteFile(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	fileID := c.Param("id")

	var file models.File
	if err := a.DB.Where("id = ? AND user_id = ?", fileID, userID).First(&file).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Delete the actual file
	if err := os.Remove(file.Path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	// Delete the database record
	if err := a.DB.Delete(&file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}
