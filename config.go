package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	tgToken string
}

func NewConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}
	return &Config{tgToken: os.Getenv("TELEGRAM_APITOKEN")}, nil
}
