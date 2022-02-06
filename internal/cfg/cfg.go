package cfg

import (
	"errors"
	"log"
	"os"
	"strconv"
)

type Config struct {
	DataBaseName               string
	BotDebug                   bool
	TelegramKey                string
	TelegramLongpullingTimeout int
	FilmLimit                  int
}

func NewConfig() (Config, error) {
	botDebug, err := GetBotDebug()
	if err != nil {
		log.Println("BOT_DEBUG wasn't set, default value false")
	}

	telegramKey, err := GetTelegramKey()
	if err != nil {
		return Config{}, err
	}

	telegramLongpullingTimeout, err := GetTelegramLongpullingTimeout()
	if err != nil {
		return Config{}, err
	}

	dbName := GetDataBaseName()

	filmLimit, err := GetAddingFilmLimit()
	if err != nil {
		return Config{}, err
	}

	return Config{
		DataBaseName:               dbName,
		BotDebug:                   botDebug,
		TelegramKey:                telegramKey,
		TelegramLongpullingTimeout: telegramLongpullingTimeout,
		FilmLimit:                  filmLimit}, nil
}

func GetBotDebug() (bool, error) {
	botDebugString := os.Getenv("BOT_DEBUG")
	botdebug, err := strconv.ParseBool(botDebugString)
	if err != nil {
		return false, &ConfigError{err}
	}
	return botdebug, nil
}

func GetTelegramKey() (string, error) {
	key := os.Getenv("TELEGRAM_KEY")
	if key == "" {
		return key, &ConfigError{errors.New("TELEGRAM_KEY is required parameter")}
	}
	return key, nil
}

func GetTelegramLongpullingTimeout() (int, error) {
	telegramLongpullingTimeoutString := os.Getenv("TELEGRAM_LONGPULLING_TIMEOUT")
	if telegramLongpullingTimeoutString == "" {
		return 60, nil
	}
	telegramLongpullingTimeout, err := strconv.ParseInt(telegramLongpullingTimeoutString, 10, 0)
	if err != nil {
		return 0, &ConfigError{err}
	}
	return int(telegramLongpullingTimeout), nil
}

func GetDataBaseName() string {
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		return "movie.db"
	}
	return dbName
}

func GetAddingFilmLimit() (int, error) {
	filmLimit := os.Getenv("FILM_LIMIT")
	if filmLimit == "" {
		return 0, nil
	}
	filmLimitInt, err := strconv.ParseInt(filmLimit, 10, 0)
	if err != nil {
		return 0, &ConfigError{err}
	}
	return int(filmLimitInt), nil
}
