package telegram

import (
	"log"

	"github.com/GeorgeS1995/moscow-movie-night-bot/internal/cfg"
	"github.com/GeorgeS1995/moscow-movie-night-bot/internal/db"
	tgbotapi "github.com/mohammadkarimi23/telegram-bot-api/v5"
)

type movieNightCommand struct {
	descriptions string
	action       func(userUpdate chan tgbotapi.Update)
}

// Proxy struct to avoid problem with tests in Send() method due to private interface Chattable
type TgBotAPIProxy struct {
	*tgbotapi.BotAPI
}

func NewTgBotAPIProxy(key string) (*TgBotAPIProxy, error) {
	bot, err := tgbotapi.NewBotAPI(key)
	if err != nil {
		return &TgBotAPIProxy{}, nil
	}
	return &TgBotAPIProxy{bot}, nil
}

func (t *TgBotAPIProxy) SendMsg(msg tgbotapi.MessageConfig) (tgbotapi.Message, error) {
	res, err := t.Send(msg)
	return res, err
}

type MovieNightTelegramBot struct {
	TGBot       BotAPIInterface
	cfg         cfg.Config
	DB          db.MovieDBInterface
	commands    map[string]movieNightCommand
	userUpdates map[int64]chan tgbotapi.Update
}

func NewMovieBot(cfg cfg.Config, db db.MovieDB) (MovieNightTelegramBot, error) {
	bot, err := NewTgBotAPIProxy(cfg.TelegramKey)
	if err != nil {
		return MovieNightTelegramBot{}, err
	}

	MovieBotIntance := MovieNightTelegramBot{TGBot: bot, cfg: cfg, DB: db}
	commands := map[string]movieNightCommand{
		"start":        {descriptions: "Приветствие от бота", action: MovieBotIntance.Greetings},
		"help":         {descriptions: "Приветствие от бота", action: MovieBotIntance.Greetings},
		"newfilm":      {descriptions: "Добавить фильм в шляпу", action: MovieBotIntance.AddFilmToHat},
		"list":         {descriptions: "Список фильмов в шляпе", action: MovieBotIntance.GetUnwatchedFilms},
		"choose":       {descriptions: "Выбрать фильм", action: MovieBotIntance.Choose},
		"list_watched": {descriptions: "Список просмотренных фильмов", action: MovieBotIntance.GetWatchedFilms},
		"editfilm":     {descriptions: "Редактировать свой добавленный фильм", action: MovieBotIntance.EditAddedFilm},
	}
	MovieBotIntance.commands = commands
	MovieBotIntance.userUpdates = make(map[int64]chan tgbotapi.Update)
	err = MovieBotIntance.setCommands(commands)
	if err != nil {
		return MovieNightTelegramBot{}, err
	}
	return MovieBotIntance, nil
}

func (b *MovieNightTelegramBot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = b.cfg.TelegramLongpullingTimeout
	u.Timeout = 60

	updates, err := b.TGBot.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		b.commandHandler(update)
		log.Println(update.Message.Text)
	}
	return nil
}

func (b *MovieNightTelegramBot) setCommands(cmd map[string]movieNightCommand) error {
	cmdList := make([]tgbotapi.BotCommand, 0, len(cmd))
	for k, v := range cmd {
		cmdList = append(cmdList, tgbotapi.BotCommand{Command: k, Description: v.descriptions})
	}
	err := b.TGBot.SetMyCommands(cmdList)
	if err != nil {
		return err
	}
	return nil
}
