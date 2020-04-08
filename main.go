package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/go-resty/resty/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/my1562/telegrambot/pkg/apiclient"
	"github.com/my1562/telegrambot/pkg/config"
	mock_apiclient "github.com/my1562/telegrambot/pkg/mockApiclient"
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
	c.Provide(func(conf *config.Config, client *resty.Client) apiclient.IApiClient {
		if conf.EmualteAPI {
			return mock_apiclient.New()
		} else {
			return apiclient.New(client)
		}
	})

	if err := c.Invoke(func(
		bot *tgbotapi.BotAPI,
		commandProcessor *CommandProcessor,
		api apiclient.IApiClient,
	) {
		log.Printf("Authorized on account %s", bot.Self.UserName)
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		updates, err := bot.GetUpdatesChan(u)
		if err != nil {
			log.Panic(err)
		}

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

			msg := tgbotapi.NewMessage(ctx.chatID, "Выберите свой адрес из списка")
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(results...)

			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		})

		commandProcessor.Callback(`subAddr:(\d+)`, func(ctx *CommandHandlerContext) {

			addressIDAr, err := strconv.ParseInt(ctx.matches[1], 10, 64)
			if err != nil {
				log.Panic(err)
			}
			log.Printf("Subscribe to: addressID:%d, chatID:%d", addressIDAr, ctx.chatID)

			err = api.CreateSubscription(ctx.chatID, addressIDAr)
			if err != nil {
				log.Panic(err)
			}

			addressString, err := api.AddressStringByID(addressIDAr)
			if err != nil {
				log.Panic(err)
			}

			msgText := "Вы подписались на обновления для адреса:\n" + addressString
			msg := tgbotapi.NewMessage(ctx.chatID, msgText)
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
