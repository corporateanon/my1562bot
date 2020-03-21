package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/go-resty/resty/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/my1562/telegrambot/pkg/apiclient"
	"github.com/my1562/telegrambot/pkg/config"
	"go.uber.org/dig"
)

func main() {

	c := dig.New()
	c.Provide(config.NewConfig)
	c.Provide(NewBotAPI)
	c.Provide(NewCommandProcessor)
	c.Provide(func(conf *config.Config) *resty.Client {
		client := resty.New().SetHostURL(conf.APIURL)
		return client
	})
	c.Provide(apiclient.New)

	if err := c.Invoke(func(
		bot *tgbotapi.BotAPI,
		commandProcessor *CommandProcessor,
		api *apiclient.ApiClient,
	) {
		log.Printf("Authorized on account %s", bot.Self.UserName)
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		updates, err := bot.GetUpdatesChan(u)
		if err != nil {
			log.Panic(err)
		}

		commandProcessor.Hears(`^hello`, func(ctx *CommandHandlerContext) {
			fmt.Println(ctx.matches)
		})

		commandProcessor.Location(func(ctx *CommandHandlerContext) {
			lat, lng := ctx.update.Message.Location.Latitude, ctx.update.Message.Location.Longitude
			geocodingResult, err := api.Geocode(lat, lng, 300)
			if err != nil {
				log.Panic(err)
			}
			addresses := geocodingResult.Addresses

			results := make([][]tgbotapi.InlineKeyboardButton, len(addresses))
			for i, address := range addresses {
				results[i] = tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						address.AddressString,
						fmt.Sprintf("subAddr:%d", address.ID),
					),
				)
			}

			msg := tgbotapi.NewMessage(ctx.chatID, "Оберіть свою адресу зі списку")
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(results...)

			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		})
		commandProcessor.Callback(`subAddr:(\d+)`, func(ctx *CommandHandlerContext) {
			fmt.Println(ctx.matches)

			addressIDAr, err := strconv.ParseInt(ctx.matches[1], 10, 64)
			if err != nil {
				panic(err)
			}

			err = api.CreateSubscription(ctx.chatID, addressIDAr)
			if err != nil {
				panic(err)
			}
		})

		for update := range updates {
			commandProcessor.Process(&update)
		}
	}); err != nil {
		log.Panic(err)
	}
}
