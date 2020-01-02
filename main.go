package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/corporateanon/my1562api"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/dig"
)

func main() {

	c := dig.New()
	c.Provide(NewConfig)
	c.Provide(NewBotAPI)
	c.Provide(NewCommandProcessor)

	if err := c.Invoke(func(
		bot *tgbotapi.BotAPI,
		commandProcessor *CommandProcessor,
	) {
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
			suggs := my1562api.GetStreetSuggestions(args.update.Message.Text)
			results := make([]string, 0)
			for index, sugg := range suggs {
				results = append(results, sugg.Name)
				if index > 10 {
					break
				}
			}
			responseText := strings.Join(results, "\n")
			if responseText == "" {
				responseText = "Nothing found"
			}
			msg := tgbotapi.NewMessage(args.update.Message.Chat.ID, responseText)

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
