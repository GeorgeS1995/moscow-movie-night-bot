package main

import (
	"log"

	"github.com/GeorgeS1995/moscow-movie-night-bot/internal/cfg"
	"github.com/GeorgeS1995/moscow-movie-night-bot/internal/db"
	tg "github.com/GeorgeS1995/moscow-movie-night-bot/internal/telegram"
)

func main() {
	cfgObj, err := cfg.NewConfig()
	if err != nil {
		log.Fatal(err)
		return
	}
	dbConn, err := db.InitMovieDB(cfgObj.DataBaseName)
	if err != nil {
		log.Fatal(err)
		return
	}
	bot, err := tg.NewMovieBot(cfgObj, dbConn)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = bot.Start()
	if err != nil {
		log.Fatal(err)
		return
	}
}
