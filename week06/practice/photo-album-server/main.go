package main

import (
	"log"
	"os"

	"photo-album-server/internal/config"
	"photo-album-server/internal/database"
	"photo-album-server/internal/handler"
	"photo-album-server/internal/router"
)

func main() {
	cfg := config.Load()

	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatalf("create upload dir failed: %v", err)
	}

	db, err := database.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("init database failed: %v", err)
	}

	h := handler.New(db, cfg)
	r := router.New(h, cfg)

	log.Printf("photo album server listening on %s", cfg.Addr)
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatalf("run server failed: %v", err)
	}
}
