package telegram

import (
	"fmt"
	"log"
)

type BotError struct {
	telegramID int64
}

type BotUserAbortError struct {
	BotError
}

func (b *BotUserAbortError) Error() string {
	msg := fmt.Sprintf("User with telegram ID %d cancel cmd", b.telegramID)
	log.Println(msg)
	return "Команда отменена"
}
