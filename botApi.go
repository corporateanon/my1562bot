package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

func NewBotAPI(conf *Config) (*tgbotapi.BotAPI, error) {
	api, err := tgbotapi.NewBotAPI(conf.tgToken)
	if err != nil {
		return nil, err
	}
	api.Debug = true

	return api, nil
}
