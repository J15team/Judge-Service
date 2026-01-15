package main

import (
	"log"

	"judge-service/internal/api"
	"judge-service/internal/config"
)

func main() {
	cfg := config.Load()

	router := api.SetupRouter(cfg)

	log.Printf("Judge Service starting on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
