package telegram

import (
	tgbotapi "github.com/mohammadkarimi23/telegram-bot-api/v5"
)

type BotAPIInterface interface {
	GetUpdatesChan(config tgbotapi.UpdateConfig) (tgbotapi.UpdatesChannel, error)
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	SetMyCommands(commands []tgbotapi.BotCommand) error
}
