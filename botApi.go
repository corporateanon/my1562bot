package main

import (
	"github.com/my1562/telegrambot/pkg/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func NewBotAPI(conf *config.Config) (*tgbotapi.BotAPI, error) {
	api, err := tgbotapi.NewBotAPI(conf.TGToken)
	if err != nil {
		return nil, err
	}
	api.Debug = true

	return api, nil
}
