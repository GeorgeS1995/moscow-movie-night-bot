package tests

import (
	"fmt"
	"testing"

	internalDB "github.com/GeorgeS1995/moscow-movie-night-bot/internal/db"
	"github.com/GeorgeS1995/moscow-movie-night-bot/internal/telegram"
	tgbotapi "github.com/mohammadkarimi23/telegram-bot-api/v5"
)

type TestTgBot struct {
	AnswerChan chan string
}

func (t *TestTgBot) GetUpdatesChan(config tgbotapi.UpdateConfig) (tgbotapi.UpdatesChannel, error) {
	return make(chan tgbotapi.Update), nil
}

func (t *TestTgBot) SendMsg(msg tgbotapi.MessageConfig) (tgbotapi.Message, error) {
	t.AnswerChan <- msg.Text
	return tgbotapi.Message{}, nil
}

func (t *TestTgBot) SetMyCommands(commands []tgbotapi.BotCommand) error {
	return nil
}

type TestDBClient struct {
	db internalDB.Movies
}

func (t *TestDBClient) SaveFilmToHat(tgUser int64, film string) error {
	return nil
}

func (t *TestDBClient) GetFilms(status internalDB.MovieStatus) (internalDB.Movies, error) {
	result := make(internalDB.Movies, 0)
	for _, m := range t.db {
		if m.Status == status {
			result = append(result, m)
		}
	}
	return result, nil
}

func (t *TestDBClient) DeleteFilmFromHat(movie string) error {
	return nil
}

func NewTestMovieBot(inmemoryDB internalDB.Movies) (telegram.MovieNightTelegramBot, chan string) {
	answerChan := make(chan string)
	return telegram.MovieNightTelegramBot{TGBot: &TestTgBot{AnswerChan: answerChan}, DB: &TestDBClient{db: inmemoryDB}}, answerChan
}

func TestChooseCMD(t *testing.T) {
	confirmAnswers := [3]string{"ДА", "Да", "да"}
	for _, confirmAnswer := range confirmAnswers {
		botAnswers := []string{".\nВы уверены, что хотите посмотреть этот фильм?\nЕсли вы ответите ДА, выбранный фильм будет удален из шляпы навсегда.\nЕсли ответите что-нибудь еще, то он останется в шляпе.", "Фильм удален из шляпы."}
		testTelegramClientInst, answerChan := NewTestMovieBot(internalDB.Movies{internalDB.Movie{Status: internalDB.MovieStatusUnwatched}, internalDB.Movie{Status: internalDB.MovieStatusUnwatched}})
		updates := make(chan tgbotapi.Update)
		go testTelegramClientInst.Choose(updates)
		updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: "", Chat: &tgbotapi.Chat{ID: 1}}}
		answer := <-answerChan
		if answer != botAnswers[0] {
			t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, botAnswers[0]))
			return
		}
		updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: confirmAnswer, Chat: &tgbotapi.Chat{ID: 1}}}
		answer = <-answerChan
		if answer != botAnswers[1] {
			t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, botAnswers[1]))
			return
		}
		t.Logf(fmt.Sprintf("TestChooseCMD complete for answer %s", confirmAnswer))
	}
}

func TestListWatchedCMD(t *testing.T) {
	inmemoryDB := internalDB.Movies{internalDB.Movie{Label: "A", Status: internalDB.MovieStatusUnwatched}, internalDB.Movie{Label: "B", Status: internalDB.MovieStatusWatched}}
	testTelegramClientInst, answerChan := NewTestMovieBot(inmemoryDB)
	updates := make(chan tgbotapi.Update)
	go testTelegramClientInst.GetWatchedFilms(updates)
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: "", Chat: &tgbotapi.Chat{ID: 1}}}
	answer := <-answerChan
	expectedAnswer := "Список просмотренных фильмов:\nB\n"
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}
	t.Logf("TestListWatchedCMD complete")
}
