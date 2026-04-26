package router

import (
	"github.com/gin-gonic/gin"

	"photo-album-server/internal/config"
	"photo-album-server/internal/handler"
	"photo-album-server/internal/middleware"
)

func New(h *handler.Handler, cfg config.AppConfig) *gin.Engine {
	r := gin.Default()
	r.MaxMultipartMemory = cfg.MaxUploadSize + (2 << 20)

	api := r.Group("/api/v1")
	{
		api.POST("/register", h.Register)
		api.POST("/login", h.Login)
	}

	protectedUploads := r.Group("/uploads")
	protectedUploads.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		protectedUploads.GET("/*filepath", h.GetProtectedFile)
	}

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		protected.POST("/albums", h.CreateAlbum)
		protected.GET("/albums", h.GetMyAlbums)
		protected.GET("/albums/public", h.GetPublicAlbums)
		protected.POST("/albums/:id/photos", h.UploadPhoto)
		protected.GET("/albums/:id/photos", h.GetAlbumPhotos)
	}

	return r
}
