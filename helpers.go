package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func wrapButtons(width int, buttons []tgbotapi.InlineKeyboardButton) [][]tgbotapi.InlineKeyboardButton {
	rowcount := len(buttons)/width + 1
	if len(buttons) == 0 {
		rowcount = 0
	}
	rows := make([][]tgbotapi.InlineKeyboardButton, rowcount)
	for i := 0; i < rowcount; i++ {
		lower := i * width
		upper := lower + width
		if upper > len(buttons) {
			upper = len(buttons)
		}
		slice := buttons[lower:upper]
		rows[i] = slice
	}
	return rows
}
