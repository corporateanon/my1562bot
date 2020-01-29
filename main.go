package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/corporateanon/my1562api"
	"github.com/corporateanon/my1562bot/pkg/config"
	"github.com/corporateanon/my1562bot/pkg/models"
	"github.com/corporateanon/my1562bot/pkg/sessionmanager"
	"github.com/corporateanon/my1562geocoder"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/gorm"
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

	if err := c.Invoke(func(
		bot *tgbotapi.BotAPI,
		commandProcessor *CommandProcessor,
		db *gorm.DB,
		sessMgr *sessionmanager.SessionManager,
		geo *my1562geocoder.Geocoder,
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

		commandProcessor.Hears(`.+`, func(ctx *CommandHandlerContext) {
			chatID := ctx.chatID
			s := sessMgr.NewSession(chatID)

			switch s.GetPhase() {
			case models.PhaseInit:
				suggs := my1562api.GetStreetSuggestions(ctx.update.Message.Text)
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
				msg := tgbotapi.NewMessage(chatID, "Нічого не знайдено")
				if len(results) > 0 {
					msg.Text = "Оберіть вулицю"
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(results...)
				}

				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}

			case models.PhaseEnterBuilding:
				streetID := s.GetStreetID()
				building := ctx.update.Message.Text

				s.SetPhase(models.PhaseInit)
				s.SetStreetID(0)
				s.Save()

				street := my1562api.GetStreetByID(streetID)
				if street == nil {
					if _, err := bot.Send(tgbotapi.NewMessage(chatID, "Error")); err != nil {
						log.Panic(err)
					}
					return
				}

				address := &models.Address{
					StreetID:   streetID,
					Building:   building,
					StreetName: street.Name,
				}

				db.Where(address).FirstOrCreate(address)

				subscription := &models.Subscription{
					ChatID: chatID,
				}
				db.Save(subscription)

				address.Subscriptions = append(address.Subscriptions, *subscription)
				db.Save(address)

				if _, err := bot.Send(
					tgbotapi.NewMessage(
						chatID,
						fmt.Sprintf("Ви обрали адресу: %s, %s", street.Name, building),
					),
				); err != nil {
					log.Panic(err)
				}
			}
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
			// var subscriptions []models.Subscription
			subscription := &models.Subscription{
				ChatID:       ctx.chatID,
				AddressIDAr:  addr.Address.ID,
				StreetID1562: addr.Street1562.ID,
			}
			db.Save(subscription)

		})

		commandProcessor.Callback(`init:`, func(ctx *CommandHandlerContext) {
			chatID := ctx.chatID

			s := sessMgr.NewSession(chatID)
			s.SetPhase(models.PhaseInit)
			s.SetStreetID(0)
			s.Save()

			msg := tgbotapi.NewMessage(chatID, "Введіть назву вулиці")
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		})

		commandProcessor.Callback(`street:(\d+)`, func(ctx *CommandHandlerContext) {
			chatID := ctx.chatID
			streetID, err := strconv.ParseInt(ctx.matches[1], 10, 64)
			if err != nil {
				log.Panic(err)
			}

			s := sessMgr.NewSession(chatID)
			s.SetPhase(models.PhaseEnterBuilding)
			s.SetStreetID(int(streetID))
			s.Save()

			msg := tgbotapi.NewMessage(chatID, "Введіть номер будинку (наприклад, 10)")
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						"...або шукати іншу вулицю",
						"init:",
					),
				),
			)
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		})

		// showSubscriptionsList := func(ctx *CommandHandlerContext) {
		// 	var subscriptions []models.Subscription
		// 	db.Where(&models.Subscription{ChatID: ctx.chatID}).Find(&subscriptions)

		// 	message := tgbotapi.NewMessage(ctx.chatID, "Підписки відсутні")

		// 	if lens := len(subscriptions); lens != 0 {
		// 		buttons := make([]tgbotapi.InlineKeyboardButton, lens)

		// 		lines := make([]string, len(subscriptions))
		// 		for i, sub := range subscriptions {
		// 			lines[i] = fmt.Sprintf("%d) %s, %s", i+1, sub.StreetName, sub.Building)
		// 			buttons[i] = tgbotapi.NewInlineKeyboardButtonData(
		// 				fmt.Sprintf("Видалити %d)", i+1),
		// 				fmt.Sprintf("subdel:%d", sub.ID),
		// 			)
		// 		}
		// 		message.Text = strings.Join(lines, "\n")
		// 		message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(wrapButtons(2, buttons)...)
		// 	}

		// 	if _, err := bot.Send(message); err != nil {
		// 		log.Panic(err)
		// 	}
		// }

		// commandProcessor.Command("list", showSubscriptionsList)
		// commandProcessor.Callback(`subdel:(\d+)`, func(ctx *CommandHandlerContext) {
		// 	chatID := ctx.chatID
		// 	id, err := strconv.ParseInt(ctx.matches[1], 10, 64)
		// 	if err != nil {
		// 		log.Panic("Could not parse")
		// 	}
		// 	sub := &models.Subscription{}
		// 	db.Where("id = ? AND chat_id = ?", id, chatID).First(sub)
		// 	db.Delete(sub)
		// 	bot.DeleteMessage(tgbotapi.NewDeleteMessage(
		// 		chatID,
		// 		ctx.update.CallbackQuery.Message.MessageID,
		// 	))
		// 	showSubscriptionsList(ctx)
		// })

		for update := range updates {
			commandProcessor.Process(&update)
		}
	}); err != nil {
		log.Panic(err)
	}
}
