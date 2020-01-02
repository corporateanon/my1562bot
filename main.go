package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/corporateanon/my1562api"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/gorm"
	"go.uber.org/dig"
)

func main() {

	c := dig.New()
	c.Provide(NewConfig)
	c.Provide(NewBotAPI)
	c.Provide(NewCommandProcessor)
	c.Provide(NewDatabase)

	if err := c.Invoke(func(
		bot *tgbotapi.BotAPI,
		commandProcessor *CommandProcessor,
		db *gorm.DB,
	) {
		defer db.Close()
		log.Printf("Authorized on account %s", bot.Self.UserName)
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		updates, err := bot.GetUpdatesChan(u)
		if err != nil {
			log.Panic(err)
		}

		commandProcessor.Hears(`^hello`, func(args *CommandHandlerArguments) {
			fmt.Println(args.matches)
		})
		commandProcessor.Hears(`.+`, func(args *CommandHandlerArguments) {
			chatID := args.update.Message.Chat.ID

			sess := Session{ChatID: chatID}
			db.Where(Session{ChatID: chatID}).FirstOrCreate(&sess)

			switch sess.Phase {
			case PhaseInit:
				suggs := my1562api.GetStreetSuggestions(args.update.Message.Text)
				results := make([][]tgbotapi.InlineKeyboardButton, 0)

				for index, sugg := range suggs {
					results = append(results,
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData(
								sugg.Name,
								fmt.Sprintf("street:%d", sugg.ID),
							)))
					if index > 10 {
						break
					}
				}
				msg := tgbotapi.NewMessage(chatID, "Nothing found")
				if len(results) > 0 {
					msg.Text = "Select your street"
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(results...)
				}

				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}

			case PhaseEnterBuilding:
				streetID := sess.StreetID
				building := args.update.Message.Text
				db.Model(&sess).Updates(map[string]interface{}{
					"Phase":    PhaseInit,
					"StreetID": 0,
				})
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Your selection: %d,%s", streetID, building))
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
			}

		})

		commandProcessor.Callback(`init:`, func(args *CommandHandlerArguments) {
			chatID := args.update.CallbackQuery.Message.Chat.ID
			sess := Session{ChatID: chatID}
			db.Where(Session{ChatID: chatID}).FirstOrCreate(&sess)
			db.Model(&sess).Updates(map[string]interface{}{
				"Phase":    PhaseInit,
				"StreetID": 0,
			})
			msg := tgbotapi.NewMessage(chatID, "Enter your street")
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		})

		commandProcessor.Callback(`street:(\d+)`, func(args *CommandHandlerArguments) {
			chatID := args.update.CallbackQuery.Message.Chat.ID
			streetID, err := strconv.ParseInt(args.matches[1], 10, 64)
			if err != nil {
				log.Panic(err)
			}

			sess := Session{ChatID: chatID}
			db.Where(Session{ChatID: chatID}).FirstOrCreate(&sess)
			db.Model(&sess).Updates(map[string]interface{}{
				"Phase":    PhaseEnterBuilding,
				"StreetID": streetID,
			})
			msg := tgbotapi.NewMessage(chatID, "Enter building number (e.g. 10)")
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						"Search other street",
						"init:",
					),
				),
			)
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		})

		for update := range updates {
			commandProcessor.Process(&update)
		}
	}); err != nil {
		log.Panic(err)
	}
}
