package utils

import (
	"github.com/joho/godotenv"
	"os"
)

func GetEnvVar(key string) string {
	err := godotenv.Load()
	if err != nil {
		return ""
	}
	return os.Getenv(key)
}
