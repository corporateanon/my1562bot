package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TGToken      string
	DBDriver     string
	DBConnection string
}

func NewConfig() (*Config, error) {
	envFile := ".env"

	injectedEnvFile := os.Getenv("ENV_FILE")
	if injectedEnvFile != "" {
		envFile = injectedEnvFile
	}

	err := godotenv.Load(envFile)
	if err != nil {
		fmt.Println(err)
	}
	return &Config{
		TGToken:      os.Getenv("TELEGRAM_APITOKEN"),
		DBDriver:     os.Getenv("DB_DRIVER"),
		DBConnection: os.Getenv("DB_CONNECTION"),
	}, nil
}
