package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TGToken    string
	APIURL     string
	EmualteAPI bool
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
		TGToken:    os.Getenv("TELEGRAM_APITOKEN"),
		APIURL:     os.Getenv("API_URL"),
		EmualteAPI: os.Getenv("EMULATE_API") == "true",
	}, nil
}
