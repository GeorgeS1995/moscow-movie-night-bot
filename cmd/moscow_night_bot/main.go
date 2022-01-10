package main

import (
	"log"

	cfg "github.com/GeorgeS1995/moscow-movie-night-bot/internal/cfg"
	"github.com/GeorgeS1995/moscow-movie-night-bot/internal/db"
	tg "github.com/GeorgeS1995/moscow-movie-night-bot/internal/telegram"
)

func main() {
	cfg, err := cfg.NewConfig()
	if err != nil {
		log.Fatal(err)
		return
	}
	db, err := db.NewMovieDB(cfg.DataBaseName)
	if err != nil {
		log.Fatal(err)
		return
	}
	bot, err := tg.NewMovieBot(cfg, db)
	if err != nil {
		log.Fatal(err)
		return
	}
	bot.Start()
}
