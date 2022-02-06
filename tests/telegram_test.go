package tests

import (
	"database/sql"
	"fmt"
	"github.com/GeorgeS1995/moscow-movie-night-bot/internal/cfg"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"regexp"
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
	ID         uint
	Label      string
	TelegramId int64
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
		testTelegramClientInst, answerChan, _ := NewTestMovieBot("./test_data/test_data.sql")
		updates := make(chan tgbotapi.Update)
		go testTelegramClientInst.Choose(updates)
		updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: "", Chat: &tgbotapi.Chat{ID: 1}}}
		answer := <-answerChan
		regexStr := "Film(1|2)\\.\nВы уверены, что хотите посмотреть этот фильм\\?\nЕсли вы ответите ДА, выбранный фильм будет удален из шляпы навсегда\\.\nЕсли ответите что-нибудь еще, то он останется в шляпе\\."
		matched, _ := regexp.MatchString(regexStr, answer)
		if !matched {
			t.Errorf(fmt.Sprintf("Not expected bot answer: %s, template: %s", answer, regexStr))
			return
		}
		updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: confirmAnswer, Chat: &tgbotapi.Chat{ID: 1}}}
		answer = <-answerChan
		expectedAnswer := "Фильм удален из шляпы."
		if answer != expectedAnswer {
			t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
			return
		}
		t.Logf(fmt.Sprintf("TestChooseCMD complete for answer %s", confirmAnswer))
	}
}

func TestListWatchedCMD(t *testing.T) {
	testTelegramClientInst, answerChan, _ := NewTestMovieBot("./test_data/test_data.sql")
	updates := make(chan tgbotapi.Update)
	go testTelegramClientInst.GetWatchedFilms(updates)
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: "", Chat: &tgbotapi.Chat{ID: 1}}}
	answer := <-answerChan
	expectedAnswer := "Список просмотренных фильмов:\nWatchedFilm\n"
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}
	t.Logf("TestListWatchedCMD complete")
}

func TestListMyFilmCMD(t *testing.T) {
	testTelegramClientInst, answerChan, _ := NewTestMovieBot("./test_data/test_data.sql")
	updates := make(chan tgbotapi.Update)
	go testTelegramClientInst.GetMyFilm(updates)
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: "", Chat: &tgbotapi.Chat{ID: 100}}}
	answer := <-answerChan
	expectedAnswer := "Список ваших фильмов на очереди:\nFilm1\n"
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}
	t.Logf("TestListMyFilmCMD complete")
}

func TestEditAddedFilmFilmNotFound(t *testing.T) {
	testTelegramClientInst, answerChan, _ := NewTestMovieBot("./test_data/test_data.sql")
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
	testTelegramClientInst, answerChan, _ := NewTestMovieBot("./test_data/test_data.sql")
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
	testTelegramClientInst, answerChan, _ := NewTestMovieBot("./test_data/test_data.sql")
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
	result := db.Raw("select movies.id, label, telegram_id from movies join users where label=@label", sql.Named("label", label)).Scan(&simpleMovie)
	if result.Error != nil {
		return simpleMovie, result.Error
	}
	return simpleMovie, nil
}

func TestEditAddedOK(t *testing.T) {
	testTelegramClientInst, answerChan, setUpConn := NewTestMovieBot("./test_data/test_data.sql")
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
	if dbResultBeforeCmdCall.TelegramId != dbResultBeforeCmdCall.TelegramId || dbResultBeforeCmdCall.ID != dbResultAfterCmdCall.ID || dbResultBeforeCmdCall.Label != searchedFims || dbResultAfterCmdCall.Label != newFimlLabel {
		t.Errorf(fmt.Sprintf("Wrong DB result, old tg_id: %d, new tg_id: %d, old id: %d, new id: %d, old label: %s, new label: %s", dbResultBeforeCmdCall.TelegramId, dbResultAfterCmdCall.TelegramId, dbResultBeforeCmdCall.ID, dbResultAfterCmdCall.ID, dbResultBeforeCmdCall.Label, dbResultAfterCmdCall.Label))
		return
	}
	t.Logf("TestEditAddedOK complete")
}

func TestEditAddedFilmCancelCMD(t *testing.T) {
	testTelegramClientInst, answerChan, setUpConn := NewTestMovieBot("./test_data/test_data.sql")
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
	if firstDbCheck.TelegramId != secondDbCheck.TelegramId || firstDbCheck.ID != secondDbCheck.ID || secondDbCheck.Label != searchedFims {
		t.Errorf(fmt.Sprintf("Wrong DB result, old tg_id: %d, new tg_id: %d, old id: %d, new id: %d, old label: %s, new label: %s", firstDbCheck.TelegramId, secondDbCheck.TelegramId, firstDbCheck.ID, secondDbCheck.ID, firstDbCheck.Label, secondDbCheck.Label))
		return
	}
	t.Logf("TestEditAddedFilmCancelCMD complete")
}

func TestAddNewFilmLimit(t *testing.T) {
	testTelegramClientInst, answerChan, _ := NewTestMovieBot("./test_data/test_data.sql")
	testTelegramClientInst.CFG = cfg.Config{FilmLimit: 1}
	updates := make(chan tgbotapi.Update)
	go testTelegramClientInst.AddFilmToHat(updates)
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: "", Chat: &tgbotapi.Chat{ID: 100}}}
	answer := <-answerChan
	expectedAnswer := "Вы привысили лимит фильмов на человека, вы сможете добавить фильм после того как какой-нибудь ваш фильм выберет шляпа. Значение лимита 1."
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}
	t.Logf("TestAddNewFilmLimit complete")
}

func TestAddNewFilmOK(t *testing.T) {
	testTelegramClientInst, answerChan, setUpConn := NewTestMovieBot("./test_data/test_data.sql")
	updates := make(chan tgbotapi.Update)
	go testTelegramClientInst.AddFilmToHat(updates)
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: "", Chat: &tgbotapi.Chat{ID: 100}}}
	answer := <-answerChan
	expectedAnswer := "Отправь мне название фильма и какую-нибудь информацию о нем (например режиссер или год), чтобы фильм опознавался однозначно."
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}
	addedFilmLabel := "TestFilm"
	updates <- tgbotapi.Update{Message: &tgbotapi.Message{Text: addedFilmLabel, Chat: &tgbotapi.Chat{ID: 100}}}
	answer = <-answerChan
	expectedAnswer = "Фильм в шляпе!"
	if answer != expectedAnswer {
		t.Errorf(fmt.Sprintf("Not expected bot answer: %s, expected: %s", answer, expectedAnswer))
		return
	}
	dbResult, err := getSimpleFilm(setUpConn, addedFilmLabel)
	if err != nil {
		t.Errorf(fmt.Sprintf("Can't get film %s from db: %s", addedFilmLabel, err.Error()))
		return
	}
	if dbResult.TelegramId != 100 {
		t.Errorf(fmt.Sprintf("Film added by not expected user. expected: %d, value: %d", 100, dbResult.TelegramId))
		return
	}
	t.Logf("TestAddNewFilmOK complete")
}
