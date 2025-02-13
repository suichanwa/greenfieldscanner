// models/file.go
package models

import (
	"time"

	"gorm.io/gorm"
)

type File struct {
	gorm.Model
	UserID       uint   `gorm:"not null"`
	Name         string `gorm:"not null"`
	Path         string `gorm:"not null"`
	Size         int64  `gorm:"not null"`
	Hash         string `gorm:"not null"`
	IsDir        bool   `gorm:"default:false"`
	ParentID     *uint
	Version      int
	LastModified time.Time
}
