package main

import (
	"log"

	"legal-riu-portal/internal/config"
	"legal-riu-portal/internal/router"
	"legal-riu-portal/internal/seed"
	"legal-riu-portal/internal/storage"
)

func main() {
	cfg := config.Load()
	db := config.InitDatabase(cfg)
	store := storage.InitMinIO(cfg)

	if err := seed.RunAllMigrationsAndSeeds(db, store); err != nil {
		log.Fatalf("Migration and seed failed: %v", err)
	}

	r, err := router.New(cfg, db, store)
	if err != nil {
		log.Fatalf("Router failed: %v", err)
	}

	log.Printf("Server running on port %s", cfg.App.Port)
	if err := r.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
