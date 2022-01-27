package telegram

import (
	"fmt"
	"log"
	"math/rand"

	internalDB "github.com/GeorgeS1995/moscow-movie-night-bot/internal/db"
	tgbotapi "github.com/mohammadkarimi23/telegram-bot-api/v5"
)

const AbortKeyboardMsg = "Отмена"

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
			_, err := b.TGBot.SendMsg(msg)
			if err != nil {
				log.Printf("Can't send msg, user: %d, msg: %s", chatID, cmd)
			}
			return
		}
		userUpdatesChannel = make(chan tgbotapi.Update)
		b.userUpdates[chatID] = userUpdatesChannel
		go handler.action(userUpdatesChannel)
		userUpdatesChannel <- update
	} else if userUpdatesExist {
		userUpdatesChannel <- update
	}
}

func (b *MovieNightTelegramBot) Greetings(updates chan tgbotapi.Update) {
	update := <-updates
	chatID := update.Message.Chat.ID
	defer delete(b.userUpdates, chatID)
	msg := tgbotapi.NewMessage(chatID, "Привет, ты можешь добавить новый фильм в шляпу с помощью команды newfilm.\nТы можешь посмотреть список фильмов в шляпе с помощью команды list.\nТы можешь выбрать фильм для просмотра с помощью комманды choose.")
	b.TGBot.SendMsg(msg)
}

func (b *MovieNightTelegramBot) AddFilmToHat(updates chan tgbotapi.Update) {
	update := <-updates
	chatID := update.Message.Chat.ID
	defer delete(b.userUpdates, chatID)
	msg := tgbotapi.NewMessage(chatID, "Отправь мне название фильма и какую-нибудь информацию о нем (например режиссер или год), чтобы фильм опознавался однозначно.")
	b.TGBot.SendMsg(msg)
	update = <-updates
	err := b.DB.SaveFilmToHat(chatID, update.Message.Text)
	if err != nil {
		msg = tgbotapi.NewMessage(chatID, "Попробуй снова, что-то пошло не так(((")
		log.Println("Can't save film: ", err)
	} else {
		msg = tgbotapi.NewMessage(chatID, "Фильм в шляпе!")
	}
	b.TGBot.SendMsg(msg)
}

func (b *MovieNightTelegramBot) GetUnwatchedFilms(updates chan tgbotapi.Update) {
	update := <-updates
	chatID := update.Message.Chat.ID
	defer delete(b.userUpdates, chatID)
	filmList, err := b.DB.GetFilms(internalDB.MovieStatusUnwatched)
	msg := tgbotapi.MessageConfig{}
	if err != nil {
		msg = tgbotapi.NewMessage(chatID, "Попробуй снова, что-то пошло не так(((")
		log.Println("Can't get film list unwatched films: ", err)
	} else {
		msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("Список фильмов на выбор:\n%s", filmList.GetMoviewList()))
	}
	b.TGBot.SendMsg(msg)
}

func (b *MovieNightTelegramBot) GetWatchedFilms(updates chan tgbotapi.Update) {
	update := <-updates
	chatID := update.Message.Chat.ID
	defer delete(b.userUpdates, chatID)
	filmList, err := b.DB.GetFilms(internalDB.MovieStatusWatched)
	msg := tgbotapi.MessageConfig{}
	if err != nil {
		msg = tgbotapi.NewMessage(chatID, "Попробуй снова, что-то пошло не так(((")
		log.Println("Can't get film list watched films: ", err)
	} else {
		msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("Список просмотренных фильмов:\n%s", filmList.GetMoviewList()))
	}
	b.TGBot.SendMsg(msg)
}

func (b *MovieNightTelegramBot) Choose(updates chan tgbotapi.Update) {
	update := <-updates
	chatID := update.Message.Chat.ID
	defer delete(b.userUpdates, chatID)
	filmList, err := b.DB.GetFilms(internalDB.MovieStatusUnwatched)
	msg := tgbotapi.MessageConfig{}
	if err != nil {
		msg = tgbotapi.NewMessage(chatID, "Попробуй снова, что-то пошло не так(((")
		log.Println("Can't get film list: ", err)
		b.TGBot.SendMsg(msg)
		return
	}
	if len(filmList) == 0 {
		msg = tgbotapi.NewMessage(chatID, "Нет фильмов в шляпе.")
		b.TGBot.SendMsg(msg)
		return
	}
	choosenFilm := filmList[rand.Intn(len(filmList))]
	msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("%s.\nВы уверены, что хотите посмотреть этот фильм?\nЕсли вы ответите ДА, выбранный фильм будет удален из шляпы навсегда.\nЕсли ответите что-нибудь еще, то он останется в шляпе.", choosenFilm.Label))
	b.TGBot.SendMsg(msg)
	update = <-updates
	answer := update.Message.Text

	_, err = ParsePositiveAnswers(answer)
	if err != nil {
		msg = tgbotapi.NewMessage(chatID, "Фильм остался в шляпе.")
	} else {
		err = b.DB.DeleteFilmFromHat(choosenFilm.Label)
		if err != nil {
			msg = tgbotapi.NewMessage(chatID, "Попробуй снова, что-то пошло не так(((")
		} else {
			msg = tgbotapi.NewMessage(chatID, "Фильм удален из шляпы.")
		}
	}
	b.TGBot.SendMsg(msg)
}

func (b *MovieNightTelegramBot) catchAnswer(updates chan tgbotapi.Update) (string, error) {
	update := <-updates
	text := update.Message.Text
	if text == AbortKeyboardMsg {
		return "", &BotUserAbortError{BotError{update.Message.Chat.ID}}
	}
	return text, nil
}

func (b *MovieNightTelegramBot) editAddedFilmDeferHandler(chatID int64, msg *tgbotapi.MessageConfig) {
	delete(b.userUpdates, chatID)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	_, err := b.TGBot.SendMsg(*msg)
	if err != nil {
		log.Printf("Can't restore keyboard for user telegram_id %d", chatID)
	}
}

func (b *MovieNightTelegramBot) EditAddedFilm(updates chan tgbotapi.Update) {
	update := <-updates
	chatID := update.Message.Chat.ID
	msg := tgbotapi.NewMessage(chatID, "Напишите название редактируемого фильма.")
	defer b.editAddedFilmDeferHandler(chatID, &msg)
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Отмена"),
		))
	b.TGBot.SendMsg(msg)
	filmLabel, err := b.catchAnswer(updates)
	if err != nil {
		msg = tgbotapi.NewMessage(chatID, err.Error())
		return
	}
	movie, err := b.DB.GetSingleFilm(internalDB.Movie{Label: filmLabel})
	if err != nil {
		msg = tgbotapi.NewMessage(chatID, "Попробуй снова, что-то пошло не так(((")
		return
	} else if movie.ID == 0 {
		msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("Фильм %s, не найден в шляпе", filmLabel))
		return
	} else if movie.Status == internalDB.MovieStatusWatched {
		msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("Фильм %s, уже в просмотренных", filmLabel))
		return
	} else if movie.TelegramID != chatID {
		msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("Фильм %s, не был добавлен вами", filmLabel))
		return
	}
	msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("Введите исправленное название для фильма %s", filmLabel))
	b.TGBot.SendMsg(msg)
	newFilmLabel, err := b.catchAnswer(updates)
	if err != nil {
		msg = tgbotapi.NewMessage(chatID, err.Error())
		return
	}

	err = b.DB.UpdateSingleFilm(movie.ID, newFilmLabel)
	if err != nil {
		msg = tgbotapi.NewMessage(chatID, err.Error())
		return
	}
	msg = tgbotapi.NewMessage(chatID, fmt.Sprintf("Фильм %s, переименован в %s", filmLabel, newFilmLabel))
}
