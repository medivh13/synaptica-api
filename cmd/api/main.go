package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	"synaptica-api/internal/app"
	"synaptica-api/internal/config"
)

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("could not load .env file: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	ctx := context.Background()

	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("startup error: %v", err)
	}

	if err := application.Run(ctx); err != nil {
		log.Fatalf("runtime error: %v", err)
	}
}
