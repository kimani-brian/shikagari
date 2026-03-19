package main

import (
	"fmt"
	"log"

	"github.com/shikagari/api/config"
	"github.com/shikagari/api/internal/router"
	"github.com/shikagari/api/pkg/database"
)

func main() {
	// ── 1. Load configuration from environment ────────────────────────────────
	cfg := config.Load()

	// ── 2. Connect to PostgreSQL and run migrations ───────────────────────────
	db := database.Connect(cfg)

	// ── 3. Initialise the Gin router with all routes registered ──────────────
	r := router.Setup(cfg, db)

	// ── 4. Start the HTTP server ──────────────────────────────────────────────
	addr := fmt.Sprintf(":%s", cfg.App.Port)
	log.Printf("[main] %s API starting on %s (env: %s)", cfg.App.Name, addr, cfg.App.Env)

	if err := r.Run(addr); err != nil {
		log.Fatalf("[main] FATAL: server failed to start: %v", err)
	}
}

// In router.go the line:
//   r.Static("/uploads", cfg.Upload.Dir)
//
// serves files from the upload directory at /uploads/*
// Example: GET /uploads/listings/1234_uuid.jpg
//
// In production, replace this with a CDN or object storage (e.g. AWS S3)
// by updating UploadService.UploadImage to write to S3 and return
// the CDN URL instead of a local path.
