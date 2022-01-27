package tests

import (
	"database/sql"
	"fmt"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"strings"
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

func migrateTestData(db *gorm.DB, pathToSql string) {
	file, _ := ioutil.ReadFile(pathToSql)
	requests := strings.Split(string(file), ";")
	for _, r := range requests {
		result := db.Exec(r)
		if result.Error != nil {
			log.Panic(result.Error)
		}
	}
}

type SimpleMovieResult struct {
	ID    uint
	Label string
}

func NewTestMovieBot(sqlTestData string) (telegram.MovieNightTelegramBot, chan string, *gorm.DB) {
	answerChan := make(chan string)
	setUpConn, _ := internalDB.OpenDB("file::memory:")
	_ = internalDB.MigrateMovieDB(setUpConn)
	migrateTestData(setUpConn, sqlTestData)
	movieBotConn := internalDB.NewMovieDB(setUpConn)
	return telegram.MovieNightTelegramBot{TGBot: &TestTgBot{AnswerChan: answerChan}, DB: movieBotConn}, answerChan, setUpConn
}

func TestChooseCMD(t *testing.T) {
	confirmAnswers := [3]string{"ДА", "Да", "да"}
	for _, confirmAnswer := range confirmAnswers {
		botAnswers := []string{"A.\nВы уверены, что хотите посмотреть этот фильм?\nЕсли вы ответите ДА, выбранный фильм будет удален из шляпы навсегда.\nЕсли ответите что-нибудь еще, то он останется в шляпе.", "Фильм удален из шляпы."}
		testTelegramClientInst, answerChan, _ := NewTestMovieBot("./test_data/one_watched_one_unwatched.sql")
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
	testTelegramClientInst, answerChan, _ := NewTestMovieBot("./test_data/one_watched_one_unwatched.sql")
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

func TestEditAddedFilmFilmNotFound(t *testing.T) {
	testTelegramClientInst, answerChan, _ := NewTestMovieBot("./test_data/editing_cmd_test_data.sql")
	updates := make(chan tgbotapi.Update)
	go testTelegramClientInst.EditAddedFilm(updates)
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: "", Chat: &tgbotapi.Chat{ID: 1}}}
	answer := <-answerChan
	expectedAnswer := "Напишите название редактируемого фильма."
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}
	searchedFims := "NotExistingFilm"
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: searchedFims, Chat: &tgbotapi.Chat{ID: 1}}}
	answer = <-answerChan
	expectedAnswer = fmt.Sprintf("Фильм %s, не найден в шляпе", searchedFims)
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}
	t.Logf("TestEditAddedFilmFilmNotFound complete")
}

func TestEditAddedFilmNotUserFilm(t *testing.T) {
	testTelegramClientInst, answerChan, _ := NewTestMovieBot("./test_data/editing_cmd_test_data.sql")
	updates := make(chan tgbotapi.Update)
	go testTelegramClientInst.EditAddedFilm(updates)
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: "", Chat: &tgbotapi.Chat{ID: 1}}}
	answer := <-answerChan
	expectedAnswer := "Напишите название редактируемого фильма."
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}
	searchedFims := "Film2"
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: searchedFims, Chat: &tgbotapi.Chat{ID: 1}}}
	answer = <-answerChan
	expectedAnswer = fmt.Sprintf("Фильм %s, не был добавлен вами", searchedFims)
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}
	t.Logf("TestEditAddedFilmNotUserFilm complete")
}

func TestEditAddedWatchedFilm(t *testing.T) {
	testTelegramClientInst, answerChan, _ := NewTestMovieBot("./test_data/editing_cmd_test_data.sql")
	updates := make(chan tgbotapi.Update)
	go testTelegramClientInst.EditAddedFilm(updates)
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: "", Chat: &tgbotapi.Chat{ID: 1}}}
	answer := <-answerChan
	expectedAnswer := "Напишите название редактируемого фильма."
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}
	searchedFims := "WatchedFilm"
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: searchedFims, Chat: &tgbotapi.Chat{ID: 1}}}
	answer = <-answerChan
	expectedAnswer = fmt.Sprintf("Фильм %s, уже в просмотренных", searchedFims)
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}
	t.Logf("TestEditAddedWatchedFilm complete")
}

func getSimpleFilm(db *gorm.DB, label string) (SimpleMovieResult, error) {
	simpleMovie := SimpleMovieResult{}
	result := db.Raw("select id, label from movies where label=@label", sql.Named("label", label)).Scan(&simpleMovie)
	if result.Error != nil {
		return simpleMovie, result.Error
	}
	return simpleMovie, nil
}

func TestEditAddedOK(t *testing.T) {
	testTelegramClientInst, answerChan, setUpConn := NewTestMovieBot("./test_data/editing_cmd_test_data.sql")
	updates := make(chan tgbotapi.Update)
	go testTelegramClientInst.EditAddedFilm(updates)
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: "", Chat: &tgbotapi.Chat{ID: 100}}}
	answer := <-answerChan
	expectedAnswer := "Напишите название редактируемого фильма."
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}
	searchedFims := "Film1"
	newFimlLabel := searchedFims + "edited"
	dbResultBeforeCmdCall, err := getSimpleFilm(setUpConn, searchedFims)
	if err != nil {
		t.Errorf(fmt.Sprintf("Can't get film %s from db: %s", searchedFims, err.Error()))
		return
	}
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: searchedFims, Chat: &tgbotapi.Chat{ID: 100}}}
	answer = <-answerChan
	expectedAnswer = fmt.Sprintf("Введите исправленное название для фильма %s", searchedFims)
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: newFimlLabel, Chat: &tgbotapi.Chat{ID: 100}}}
	answer = <-answerChan
	expectedAnswer = fmt.Sprintf("Фильм %s, переименован в %s", searchedFims, newFimlLabel)
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}

	dbResultAfterCmdCall, err := getSimpleFilm(setUpConn, newFimlLabel)
	if err != nil {
		t.Errorf(fmt.Sprintf("Can't get film %s from db: %s", newFimlLabel, err.Error()))
		return
	}
	if dbResultBeforeCmdCall.ID != dbResultAfterCmdCall.ID || dbResultBeforeCmdCall.Label != searchedFims || dbResultAfterCmdCall.Label != newFimlLabel {
		t.Errorf(fmt.Sprintf("Wrong DB result, old id: %d, new id: %d, old label: %s, new label: %s", dbResultBeforeCmdCall.ID, dbResultAfterCmdCall.ID, dbResultBeforeCmdCall.Label, dbResultAfterCmdCall.Label))
		return
	}
	t.Logf("TestEditAddedOK complete")
}

func TestEditAddedFilmCancelCMD(t *testing.T) {
	testTelegramClientInst, answerChan, setUpConn := NewTestMovieBot("./test_data/editing_cmd_test_data.sql")
	updates := make(chan tgbotapi.Update)
	go testTelegramClientInst.EditAddedFilm(updates)
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: "", Chat: &tgbotapi.Chat{ID: 100}}}
	answer := <-answerChan
	expectedAnswer := "Напишите название редактируемого фильма."
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}
	searchedFims := "Film1"
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: searchedFims, Chat: &tgbotapi.Chat{ID: 100}}}
	answer = <-answerChan
	expectedAnswer = fmt.Sprintf("Введите исправленное название для фильма %s", searchedFims)
	firstDbCheck, err := getSimpleFilm(setUpConn, searchedFims)
	if err != nil {
		t.Errorf(fmt.Sprintf("Can't get film %s from db: %s", searchedFims, err.Error()))
		return
	}
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: "Отмена", Chat: &tgbotapi.Chat{ID: 100}}}
	answer = <-answerChan
	expectedAnswer = "Команда отменена"
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}
	secondDbCheck, err := getSimpleFilm(setUpConn, searchedFims)
	if err != nil {
		t.Errorf(fmt.Sprintf("Can't get film %s from db: %s", searchedFims, err.Error()))
		return
	}
	if firstDbCheck.ID != secondDbCheck.ID || secondDbCheck.Label != searchedFims {
		t.Errorf(fmt.Sprintf("Wrong DB result, old id: %d, new id: %d, old label: %s, new label: %s", firstDbCheck.ID, secondDbCheck.ID, firstDbCheck.Label, secondDbCheck.Label))
		return
	}
	t.Logf("TestEditAddedFilmCancelCMD complete")
}
