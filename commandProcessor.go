package main

import (
	"regexp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type RuleType uint8

const (
	Hears RuleType = iota
	Callback
	Command
	Location
)

type CommandHandlerContext struct {
	update  *tgbotapi.Update
	chatID  int64
	matches []string
}
type CommandHandler func(update *CommandHandlerContext)

type CommandProcessorRule struct {
	re       *regexp.Regexp
	handler  CommandHandler
	ruleType RuleType
}

type CommandProcessor struct {
	api   *tgbotapi.BotAPI
	rules []*CommandProcessorRule
}

func (cp *CommandProcessor) Hears(reStr string, handler CommandHandler) {
	re := regexp.MustCompile(reStr)
	cp.rules = append(cp.rules, &CommandProcessorRule{re: re, handler: handler, ruleType: Hears})
}

func (cp *CommandProcessor) Callback(reStr string, handler CommandHandler) {
	re := regexp.MustCompile(reStr)
	cp.rules = append(cp.rules, &CommandProcessorRule{re: re, handler: handler, ruleType: Callback})
}

func (cp *CommandProcessor) Command(reStr string, handler CommandHandler) {
	re := regexp.MustCompile(reStr)
	cp.rules = append(cp.rules, &CommandProcessorRule{re: re, handler: handler, ruleType: Command})
}

func (cp *CommandProcessor) Location(handler CommandHandler) {
	cp.rules = append(cp.rules, &CommandProcessorRule{handler: handler, ruleType: Location})
}

func (cp *CommandProcessor) Process(update *tgbotapi.Update) {
	for _, rule := range cp.rules {

		var data string
		var chatID int64
		switch rule.ruleType {
		case Callback:
			if update.CallbackQuery == nil {
				continue
			}
			data = update.CallbackQuery.Data
			chatID = update.CallbackQuery.Message.Chat.ID

		case Hears:
			if update.Message == nil || update.Message.IsCommand() {
				continue
			}
			data = update.Message.Text
			chatID = update.Message.Chat.ID

		case Command:
			if update.Message == nil || !update.Message.IsCommand() {
				continue
			}
			data = update.Message.Command()
			chatID = update.Message.Chat.ID

		case Location:
			if update.Message == nil || update.Message.Location == nil {
				continue
			}
			chatID = update.Message.Chat.ID
			go rule.handler(&CommandHandlerContext{
				update: update,
				chatID: chatID,
			})
			continue
		}

		submatches := rule.re.FindStringSubmatch(data)

		if len(submatches) > 0 {
			go rule.handler(&CommandHandlerContext{
				update:  update,
				matches: submatches,
				chatID:  chatID,
			})
			break
		}
	}
}

func NewCommandProcessor(api *tgbotapi.BotAPI) *CommandProcessor {
	return &CommandProcessor{api: api, rules: []*CommandProcessorRule{}}
}
