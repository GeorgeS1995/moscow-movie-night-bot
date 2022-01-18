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

type MovieNightTelegramBot struct {
	TGBot       BotAPIInterface
	cfg         cfg.Config
	DB          db.MovieDBInterface
	commands    map[string]movieNightCommand
	userUpdates map[int64]chan tgbotapi.Update
}

func NewMovieBot(cfg cfg.Config, db db.MovieDB) (MovieNightTelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramKey)
	if err != nil {
		return MovieNightTelegramBot{}, err
	}

	MovieBotIntance := MovieNightTelegramBot{TGBot: bot, cfg: cfg, DB: db}
	commands := map[string]movieNightCommand{
		"start":   {descriptions: "Приветствие от бота", action: MovieBotIntance.Greetings},
		"help":    {descriptions: "Приветствие от бота", action: MovieBotIntance.Greetings},
		"newfilm": {descriptions: "Добавить фильм в шляпу", action: MovieBotIntance.AddFilmToHat},
		"list":    {descriptions: "Список фильмов в шляпе", action: MovieBotIntance.GetAllFilms},
		"choose":  {descriptions: "Выбрать фильм", action: MovieBotIntance.Choose},
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
