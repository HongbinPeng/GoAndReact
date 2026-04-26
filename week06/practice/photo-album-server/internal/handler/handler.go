package handler

import (
	"errors"

	"photo-album-server/internal/config"

	"gorm.io/gorm"
)

var (
	errFileNotFound    = errors.New("file not found")
	errFileForbidden   = errors.New("file forbidden")
	errInvalidFilePath = errors.New("invalid file path")
	errQueryFileFailed = errors.New("query file failed")
)

type Handler struct {
	DB     *gorm.DB
	Config config.AppConfig
}

func New(db *gorm.DB, cfg config.AppConfig) *Handler {
	return &Handler{
		DB:     db,
		Config: cfg,
	}
}
