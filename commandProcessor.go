package main

import (
	"regexp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type CommandHandlerArguments struct {
	update  *tgbotapi.Update
	matches []string
}
type CommandHandler func(update *CommandHandlerArguments)

type CommandProcessorRule struct {
	re         *regexp.Regexp
	handler    CommandHandler
	isCallback bool
}

type CommandProcessor struct {
	api   *tgbotapi.BotAPI
	rules []*CommandProcessorRule
}

func (cp *CommandProcessor) Hears(reStr string, handler CommandHandler) {
	re := regexp.MustCompile(reStr)
	cp.rules = append(cp.rules, &CommandProcessorRule{re: re, handler: handler, isCallback: false})
}

func (cp *CommandProcessor) Callback(reStr string, handler CommandHandler) {
	re := regexp.MustCompile(reStr)
	cp.rules = append(cp.rules, &CommandProcessorRule{re: re, handler: handler, isCallback: true})
}

func (cp *CommandProcessor) Process(update *tgbotapi.Update) {
	for _, rule := range cp.rules {

		var data string
		if rule.isCallback {
			if update.CallbackQuery == nil {
				continue
			}
			data = update.CallbackQuery.Data
		} else {
			if update.Message == nil {
				continue
			}
			data = update.Message.Text
		}

		submatches := rule.re.FindStringSubmatch(data)

		if len(submatches) > 0 {
			go rule.handler(&CommandHandlerArguments{
				update:  update,
				matches: submatches,
			})
			break
		}
	}
}

func NewCommandProcessor(api *tgbotapi.BotAPI) *CommandProcessor {
	return &CommandProcessor{api: api, rules: []*CommandProcessorRule{}}
}
