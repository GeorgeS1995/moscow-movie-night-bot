package telegram

import (
	"fmt"
	"log"
	"math/rand"

	tgbotapi "github.com/mohammadkarimi23/telegram-bot-api/v5"
)

func (b *MovieNightTelegramBot) commandHandler(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	isCmd := update.Message.IsCommand()

	// Get and create if needed individual command update channel
	userUpdatesChannel, userUpdatesExist := b.userUpdates[chatID]
	if !userUpdatesExist && isCmd {
		cmd := update.Message.Command()
		handler, handlerExist := b.commands[cmd]
		if !handlerExist {
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Я не знаю команды: %v", cmd))
			_, err := b.TGBot.Send(msg)
			if err != nil {
				log.Printf("Can't send msg, user: %d, msg: %s", chatID, cmd)
			}
			return
		}
		userUpdatesChannel = make(chan tgbotapi.Update)
		b.userUpdates[chatID] = userUpdatesChannel
		go handler.action(userUpdatesChannel)
		userUpdatesChannel <- update
	} else if userUpdatesExist && isCmd {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Завершите предыдущую команду прежде чем вызывать текущую"))
		_, err := b.TGBot.Send(msg)
		log.Printf("Attempt to call commad during another command: %d, msg: %s", chatID, update.Message.Text)
		if err != nil {
			log.Printf("Can't send msg, user: %d, msg: %s", chatID, update.Message.Text)
		}
	} else if userUpdatesExist {
		userUpdatesChannel <- update
	}
}

func (b *MovieNightTelegramBot) Greetings(updates chan tgbotapi.Update) {
	update := <-updates
	chatID := update.Message.Chat.ID
	defer delete(b.userUpdates, chatID)
	msg := tgbotapi.NewMessage(chatID, "Привет, ты можешь добавить новый фильм в шляпу с помощью команды newfilm.\nТы можешь посмотреть список фильмов в шляпе с помощью команды list.\nТы можешь выбрать фильм для просмотра с помощью комманды choose.")
	b.TGBot.Send(msg)
}

func (b *MovieNightTelegramBot) AddFilmToHat(updates chan tgbotapi.Update) {
	update := <-updates
	chatID := update.Message.Chat.ID
	defer delete(b.userUpdates, chatID)
	msg := tgbotapi.NewMessage(chatID, "Отправь мне название фильма и какую-нибудь информацию о нем (например режиссер или год), чтобы фильм опознавался однозначно.")
	b.TGBot.Send(msg)
	update = <-updates
	err := b.DB.SaveFilmToHat(chatID, update.Message.Text)
	if err != nil {
		msg = tgbotapi.NewMessage(chatID, "Попробуй снова, что-то пошло не так(((")
		log.Println("Can't save film: ", err)
	} else {
		msg = tgbotapi.NewMessage(chatID, "Фильм в шляпе!")
	}
	b.TGBot.Send(msg)
}

func (b *MovieNightTelegramBot) GetAllFilms(updates chan tgbotapi.Update) {
	update := <-updates
	chatID := update.Message.Chat.ID
	defer delete(b.userUpdates, chatID)
	filmList, err := b.DB.GetAllFilms()
	msg := tgbotapi.MessageConfig{}
	if err != nil {
		msg = tgbotapi.NewMessage(chatID, "Попробуй снова, что-то пошло не так(((")
		log.Println("Can't get film list: ", err)
	} else {
		msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("Список фильмов на выбор:\n%s", filmList.GetMoviewList()))
	}
	b.TGBot.Send(msg)
}

func (b *MovieNightTelegramBot) Choose(updates chan tgbotapi.Update) {
	update := <-updates
	chatID := update.Message.Chat.ID
	defer delete(b.userUpdates, chatID)
	filmList, err := b.DB.GetAllFilms()
	msg := tgbotapi.MessageConfig{}
	if err != nil {
		msg = tgbotapi.NewMessage(chatID, "Попробуй снова, что-то пошло не так(((")
		log.Println("Can't get film list: ", err)
		b.TGBot.Send(msg)
		return
	}
	if len(filmList) == 0 {
		msg = tgbotapi.NewMessage(chatID, "Нет фильмов в шляпе.")
		b.TGBot.Send(msg)
		return
	}
	choosenFilm := filmList[rand.Intn(len(filmList))]
	msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("%s.\nВы уверены, что хотите посмотреть этот фильм?\nЕсли вы ответите ДА, выбранный фильм будет удален из шляпы навсегда.\nЕсли ответите что-нибудь еще, то он останется в шляпе.", choosenFilm.Label))
	b.TGBot.Send(msg)
	update = <-updates
	answer := update.Message.Text

	_, err = ParsePositiveAnswers(answer)
	if err != nil {
		msg = tgbotapi.NewMessage(chatID, fmt.Sprint("Фильм остался в шляпе."))
	} else {
		err = b.DB.DeleteFilmFromHat(choosenFilm.Label)
		if err != nil {
			msg = tgbotapi.NewMessage(chatID, fmt.Sprint("Попробуй снова, что-то пошло не так((("))
		} else {
			msg = tgbotapi.NewMessage(chatID, fmt.Sprint("Фильм удален из шляпы."))
		}
	}
	b.TGBot.Send(msg)
}
