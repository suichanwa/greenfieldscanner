// main.go
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type App struct {
	DB     *gorm.DB
	Router *gin.Engine
}

func main() {
	app := App{}
	app.Initialize("storage.db") // SQLite database file
	app.Run(":8080")
}

func (a *App) Initialize(dbName string) {
	a.Router.POST("/api/v1/register", a.Register)
	a.Router.POST("/api/v1/login", a.Login)

	authGroup := a.Router.Group("/api/v1").Use(a.JWTAuthMiddleware())
	{
		authGroup.POST("/upload", a.UploadFile)
		authGroup.GET("/files", a.ListFiles)
		authGroup.POST("/sync", a.Sync)
		authGroup.GET("/files/:id/download", a.DownloadFile)
	}

	var err error
	a.DB, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate models
	a.DB.AutoMigrate(&User{}, &File{})

	// Initialize router
	a.Router = gin.Default()
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	log.Println("Server running on", addr)
	if err := a.Router.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
