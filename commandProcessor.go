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
	re      *regexp.Regexp
	handler CommandHandler
}

type CommandProcessor struct {
	api   *tgbotapi.BotAPI
	rules []*CommandProcessorRule
}

func (cp *CommandProcessor) Hears(reStr string, handler CommandHandler) {
	re := regexp.MustCompile(reStr)
	cp.rules = append(cp.rules, &CommandProcessorRule{re: re, handler: handler})
}

func (cp *CommandProcessor) Process(update *tgbotapi.Update) {
	for _, rule := range cp.rules {
		if update.Message == nil {
			return
		}
		submatches := rule.re.FindStringSubmatch(update.Message.Text)
		if len(submatches) > 0 {
			rule.handler(&CommandHandlerArguments{
				update:  update,
				matches: submatches,
			})
		}
	}
}

func NewCommandProcessor(api *tgbotapi.BotAPI) *CommandProcessor {
	return &CommandProcessor{api: api, rules: []*CommandProcessorRule{}}
}
