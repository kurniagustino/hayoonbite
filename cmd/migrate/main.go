package main

import (
	"hayoon-bite-backend/internal/database"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Load env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// 2. Connect Database
	database.Connect()

	// 3. Jalankan Migrasi
	database.Migrate()
}
