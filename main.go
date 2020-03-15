package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/go-resty/resty/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/gorm"
	"github.com/my1562/geocoder"
	"github.com/my1562/telegrambot/pkg/apiclient"
	"github.com/my1562/telegrambot/pkg/config"
	"github.com/my1562/telegrambot/pkg/models"
	"github.com/my1562/telegrambot/pkg/sessionmanager"
	"go.uber.org/dig"
)

func main() {

	c := dig.New()
	c.Provide(config.NewConfig)
	c.Provide(NewBotAPI)
	c.Provide(NewCommandProcessor)
	c.Provide(models.NewDatabase)
	c.Provide(sessionmanager.NewSessionManager)
	c.Provide(NewGeocoder)
	c.Provide(func(conf *config.Config) *resty.Client {
		client := resty.New().SetHostURL(conf.APIURL)
		return client
	})
	c.Provide(apiclient.New)

	if err := c.Invoke(func(
		bot *tgbotapi.BotAPI,
		commandProcessor *CommandProcessor,
		db *gorm.DB,
		sessMgr *sessionmanager.SessionManager,
		geo *geocoder.Geocoder,
		api *apiclient.ApiClient,
	) {
		defer db.Close()
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
			geocodingResults := geo.ReverseGeocode(lat, lng, 300, 10)

			results := make([][]tgbotapi.InlineKeyboardButton, len(geocodingResults))
			for i, geoRes := range geocodingResults {
				results[i] = tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						FormatGeocodingResult(geoRes),
						fmt.Sprintf("subAddr:%d", geoRes.FullAddress.Address.ID),
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

			addressIDAr, err := strconv.ParseUint(ctx.matches[1], 10, 64)
			if err != nil {
				panic(err)
			}

			addr := geo.AddressByID(uint32(addressIDAr))
			if addr == nil {
				panic("No such address") //TODO: send it as a message
			}

			err = api.CreateSubscription(ctx.chatID, int64(addr.Address.ID))
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
