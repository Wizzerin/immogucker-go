package config

import (
	"log"

	"github.com/joho/godotenv"
)

// Load initializes environment variables from a .env file if it exists
func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found. Falling back to system environment variables.")
	}
}
