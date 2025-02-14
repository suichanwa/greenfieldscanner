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

func (a *App) ListFiles(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var files []models.File
	if err := a.DB.Where("user_id = ?", userID).Find(&files).Error; err != nil {
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

func (a *App) Sync(c *gin.Context) {
	type SyncRequest struct {
		Files []struct {
			Name         string    `json:"name"`
			Hash         string    `json:"hash"`
			LastModified time.Time `json:"last_modified"`
		} `json:"files"`
	}

	var req SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID := c.MustGet("userID").(uint)

	var changes struct {
		ToUpload   []string      `json:"to_upload"`
		ToDownload []models.File `json:"to_download"`
	}

	// Check each client file against server
	for _, clientFile := range req.Files {
		var serverFile models.File
		err := a.DB.Where("user_id = ? AND name = ?", userID, clientFile.Name).First(&serverFile).Error

		if err != nil {
			// File doesn't exist on server
			changes.ToUpload = append(changes.ToUpload, clientFile.Name)
			continue
		}

		if serverFile.Hash != clientFile.Hash {
			if serverFile.LastModified.After(clientFile.LastModified) {
				changes.ToDownload = append(changes.ToDownload, serverFile)
			} else {
				changes.ToUpload = append(changes.ToUpload, clientFile.Name)
			}
		}
	}

	c.JSON(http.StatusOK, changes)
}
