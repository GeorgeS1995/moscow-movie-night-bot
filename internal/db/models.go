package db

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// BaseInheritedModel Simplifiest gorm.Model struct version
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

// SingleMovie Returning by GetSingleFilm func
type SingleMovie struct {
	ID         uint
	Label      string
	Status     MovieStatus
	TelegramID int64
}

type Movies []Movie

func (movies Movies) GetMoviesList(randomSeq bool) string {
	var movieListBuilder strings.Builder
	for {
		if len(movies) == 0 {
			break
		}
		var choose int
		if randomSeq {
			choose = rand.Intn(len(movies))
		}
		movieListBuilder.WriteString(fmt.Sprintln(movies[choose].Label))
		movies = append(movies[:choose], movies[choose+1:]...)
	}
	return movieListBuilder.String()
}

type User struct {
	BaseInheritedModel
	TelegramID int64 `gorm:"unique;"`
}

type MovieSearch struct {
	Status     *MovieStatus
	TelegramID *int64
}

func and(s string) string {
	if s != "" {
		s += " AND "
	}
	return s
}

// GetSearchQuery Returns gorm query string for Where clause and placeholders args
func (ms *MovieSearch) GetSearchQuery() (string, []interface{}) {
	var query string
	var args []interface{}
	if ms.Status != nil {
		query += "status = ?"
		args = append(args, *ms.Status)
	}
	if ms.TelegramID != nil {
		query = and(query)
		query += "telegram_id = ?"
		args = append(args, *ms.TelegramID)
	}
	return query, args
}
