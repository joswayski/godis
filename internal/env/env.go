package env

import (
	"log/slog"

	"github.com/joho/godotenv"
)

func GetEnv() {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("Unable to get .env file!")
	}
}
