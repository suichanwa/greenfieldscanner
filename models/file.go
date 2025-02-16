package models

import (
	"time"

	"gorm.io/gorm"
)

type File struct {
	gorm.Model
	UserID       uint      `json:"user_id" gorm:"not null"`
	Name         string    `json:"name" gorm:"not null"`
	Path         string    `json:"path" gorm:"not null"`
	Size         int64     `json:"size" gorm:"not null"`
	Hash         string    `json:"hash" gorm:"not null;index"`
	IsDir        bool      `json:"is_dir" gorm:"default:false"`
	ParentID     *uint     `json:"parent_id"`
	Version      int       `json:"version" gorm:"default:1"`
	LastModified time.Time `json:"last_modified"`
}
