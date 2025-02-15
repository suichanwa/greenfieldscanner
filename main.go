package main

import (
	"cloud-storage/handlers"
	"cloud-storage/middleware"
	"cloud-storage/models"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type App struct {
	handlers.App
}

func main() {
	app := App{}
	app.Initialize("storage.db")
	app.Run(":8080")
}

func (a *App) Initialize(dbName string) {
	a.Router = gin.Default()

	var err error
	a.DB, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	a.App = handlers.App{
		DB:     a.DB,
		Router: a.Router,
	}

	a.DB.AutoMigrate(&models.User{}, &models.File{})

	if err := os.MkdirAll("storage", 0755); err != nil {
		log.Fatal("Failed to create storage directory:", err)
	}

	a.Router.POST("/api/v1/register", a.Register)
	a.Router.POST("/api/v1/login", a.Login)

	authGroup := a.Router.Group("/api/v1").Use(middleware.JWTAuthMiddleware())
	{
		authGroup.POST("/upload", a.UploadFile)
		authGroup.GET("/files", a.ListFiles)
		authGroup.POST("/sync", a.Sync)
		authGroup.GET("/files/:id/download", a.DownloadFile)
	}
}

func (a *App) Run(addr string) {
	log.Printf("Server running on %s\nEndpoints:\n"+
		"POST /api/v1/register - Register new user\n"+
		"POST /api/v1/login - Login\n"+
		"POST /api/v1/upload - Upload file (requires auth)\n"+
		"GET /api/v1/files - List files (requires auth)\n"+
		"GET /api/v1/files/:id/download - Download file (requires auth)\n"+
		"POST /api/v1/sync - Sync files (requires auth)\n", addr)
	if err := a.Router.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
