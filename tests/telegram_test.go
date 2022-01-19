package tests

import (
	"fmt"
	"testing"

	internalDB "github.com/GeorgeS1995/moscow-movie-night-bot/internal/db"
	"github.com/GeorgeS1995/moscow-movie-night-bot/internal/telegram"
	tgbotapi "github.com/mohammadkarimi23/telegram-bot-api/v5"
)

type TestTgBot struct {
	answers    []string
	testClient *testing.T
}

func (t *TestTgBot) GetUpdatesChan(config tgbotapi.UpdateConfig) (tgbotapi.UpdatesChannel, error) {
	return make(chan tgbotapi.Update), nil
}

func (t *TestTgBot) SendMsg(msg tgbotapi.MessageConfig) (tgbotapi.Message, error) {
	if msg.Text != t.answers[0] {
		t.testClient.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", msg.Text, t.answers[0]))
	}
	t.answers = t.answers[1:]
	return tgbotapi.Message{}, nil
}

func (t *TestTgBot) SetMyCommands(commands []tgbotapi.BotCommand) error {
	return nil
}

type TestDBClient struct{}

func (t *TestDBClient) SaveFilmToHat(tgUser int64, film string) error {
	return nil
}

func (t *TestDBClient) GetAllFilms() (internalDB.Movies, error) {
	return internalDB.Movies{internalDB.Movie{}, internalDB.Movie{}}, nil
}

func (t *TestDBClient) DeleteFilmFromHat(movie string) error {
	return nil
}

func NewTestMovieBot(botAnswers []string, testClient *testing.T) telegram.MovieNightTelegramBot {
	return telegram.MovieNightTelegramBot{TGBot: &TestTgBot{answers: botAnswers}, DB: &TestDBClient{}}
}

func TestChooseCMD(t *testing.T) {
	confirmAnswers := [3]string{"ДА", "Да", "да"}
	for _, confirmAnswer := range confirmAnswers {
		botAnswers := []string{".\nВы уверены, что хотите посмотреть этот фильм?\nЕсли вы ответите ДА, выбранный фильм будет удален из шляпы навсегда.\nЕсли ответите что-нибудь еще, то он останется в шляпе.", "Фильм удален из шляпы."}
		testTelegramClientInst := NewTestMovieBot(botAnswers, t)
		updates := make(chan tgbotapi.Update)
		go testTelegramClientInst.Choose(updates)
		updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: "", Chat: &tgbotapi.Chat{ID: 1}}}
		updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: confirmAnswer, Chat: &tgbotapi.Chat{ID: 1}}}
		t.Logf(fmt.Sprintf("TestChooseCMD complete for answer %s", confirmAnswer))
	}
}
