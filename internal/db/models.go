package db

import (
	"fmt"
	"math/rand"
	"strings"

	gorm "gorm.io/gorm"
)

type Movie struct {
	gorm.Model
	Label string
	Users []User `gorm:"many2many:user_movies;"`
}

type Movies []Movie

func (movies Movies) GetMoviewList() string {
	var movieListBuilder strings.Builder
	for {
		if len(movies) == 1 {
			movieListBuilder.WriteString(fmt.Sprintln(movies[0].Label))
			break
		}
		choose := rand.Intn(len(movies))
		movieListBuilder.WriteString(fmt.Sprintln(movies[choose].Label))
		movies = append(movies[:choose], movies[choose+1:]...)
	}
	return movieListBuilder.String()
}

type User struct {
	gorm.Model
	TelegramID int64   `gorm:"unique;"`
	Movies     []Movie `gorm:"many2many:user_movies;"`
}
