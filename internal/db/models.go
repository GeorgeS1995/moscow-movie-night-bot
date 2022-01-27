package db

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Simplifiest gorm.Model struct version
type BaseInheritedModel struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Movie struct {
	BaseInheritedModel
	Label  string
	UserID uint `gorm:"foreignKey:UserID;references:ID"`
	Status MovieStatus
}

// Returning by GetSingleFilm func
type SingleMovie struct {
	ID         uint
	Label      string
	Status     MovieStatus
	TelegramID int64
}

type Movies []Movie

func (movies Movies) GetMoviewList() string {
	var movieListBuilder strings.Builder
	for {
		if len(movies) == 0 {
			break
		}
		choose := rand.Intn(len(movies))
		movieListBuilder.WriteString(fmt.Sprintln(movies[choose].Label))
		movies = append(movies[:choose], movies[choose+1:]...)
	}
	return movieListBuilder.String()
}

type User struct {
	BaseInheritedModel
	TelegramID int64 `gorm:"unique;"`
}
