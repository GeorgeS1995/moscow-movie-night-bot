package db

import (
	"fmt"
	"strings"

	gorm "gorm.io/gorm"
)

type Movie struct {
	gorm.Model
	Label string
	Users []User `gorm:"many2many:user_movies;"`
}

type Movies []Movie

func (movies *Movies) GetMoviewList() string {
	var movieListBuilder strings.Builder
	for _, m := range *movies {
		movieListBuilder.WriteString(fmt.Sprintln(m.Label))
	}
	return movieListBuilder.String()
}

type User struct {
	gorm.Model
	TelegramID int64   `gorm:"unique;"`
	Movies     []Movie `gorm:"many2many:user_movies;"`
}
