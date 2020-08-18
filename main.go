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
	"github.com/my1562/telegrambot/pkg/priorityChecker"
	"go.uber.org/dig"
)

func getAddressStatusMessage(
	message string,
	addressString string,
	addressStatus apiclient.AddressArCheckStatus,
) string {

	introduction := ""
	emojiIcon := ""

	if addressStatus == apiclient.AddressStatusNoWork {
		introduction = "Работы не проводятся"
		emojiIcon = "✅"
	}
	if addressStatus == apiclient.AddressStatusWork {
		introduction = ""
		emojiIcon = "🛠"
	}

	fullMessageText := emojiIcon + " " + addressString + ": " + introduction + "\n\n" + message
	return fullMessageText
}

func main() {

	c := dig.New()
	c.Provide(config.NewConfig)
	c.Provide(NewBotAPI)
	c.Provide(NewCommandProcessor)
	c.Provide(func(conf *config.Config) *resty.Client {
		client := resty.New().SetHostURL(conf.APIURL)
		return client
	}, dig.Name("api"))
	c.Provide(func(conf *config.Config) *resty.Client {
		client := resty.New().SetHostURL(conf.FTSURL)
		return client
	}, dig.Name("fts"))

	c.Provide(func(conf *config.Config, clientOptions apiclient.ApiClientOptions) apiclient.IApiClient {
		if conf.EmualteAPI {
			return mock_apiclient.New()
		}
		return apiclient.New(clientOptions)
	})

	c.Provide(priorityChecker.NewPriorityChecker)

	if err := c.Invoke(func(
		bot *tgbotapi.BotAPI,
		commandProcessor *CommandProcessor,
		api apiclient.IApiClient,
		priorityChecker *priorityChecker.PriorityChecker,
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

			addressAr, err := api.AddressByID(addressIDAr)
			if err != nil {
				log.Panic(err)
			}
			if addressAr != nil && addressAr.CheckStatus != apiclient.AddressStatusInit {
				msgText := getAddressStatusMessage(addressAr.ServiceMessage, addressString, addressAr.CheckStatus)
				msg := tgbotapi.NewMessage(ctx.chatID, msgText)
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
			}

			msgText := "Вы подписались на обновления для адреса: " + addressString + ".\nКак только появится новая информация по вашему адресу, мы отправим вам сообщение"
			msg := tgbotapi.NewMessage(ctx.chatID, msgText)
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}

			log.Printf("Enqueueing priority check for address (ID=%d): %s\n", addressIDAr, addressString)
			if err := priorityChecker.EnqueuePriorityCheck(addressIDAr); err != nil {
				log.Panic(err)
			}
		})

		commandProcessor.Hears(".*", func(ctx *CommandHandlerContext) {
			if len(ctx.matches) == 0 {
				return
			}
			query := ctx.matches[0]

			addresses, err := api.FullTextSearch(query)
			if err != nil {
				log.Panic(err)
			}
			if len(addresses) == 0 {
				msgText := "Ничего не найдено.\nПопробуйте иначе сформулировать запрос.\nНапример:\n\nмироносицкая 41"
				msg := tgbotapi.NewMessage(ctx.chatID, msgText)
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
				return
			}

			results := make([][]tgbotapi.InlineKeyboardButton, len(addresses))
			for i, address := range addresses {
				results[i] = tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						address.Label,
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

		commandProcessor.Command("start", func(ctx *CommandHandlerContext) {
			msg := tgbotapi.NewMessage(ctx.chatID, "Чтобы начать, отправьте боту свою геолокацию или введите свой адрес (улица, дом)")

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
