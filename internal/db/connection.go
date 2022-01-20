package db

import (
	"errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MovieDB struct {
	db *gorm.DB
}

func NewMovieDB(dbname string) (MovieDB, error) {
	db, err := gorm.Open(sqlite.Open(dbname), &gorm.Config{})
	if err != nil {
		return MovieDB{}, err
	}
	db.AutoMigrate(&Movie{}, &User{})
	return MovieDB{db: db}, nil
}

func (m MovieDB) GetOrCreateUser(tgUser int64) (User, error) {
	user := User{}
	result := m.db.Where("telegram_id = ?", tgUser).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			user = User{TelegramID: tgUser}
			result = m.db.Create(&user)
		}
		return user, result.Error
	}
	return user, nil
}

func (m MovieDB) GetOrCreateMovie(film string) (Movie, error) {
	movie := Movie{}
	result := m.db.Where("label = ?", film).First(&movie)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			movie = Movie{Label: film, Status: MovieStatusUnwatched}
			result = m.db.Create(&movie)
		}
		return movie, result.Error
	}
	return movie, nil
}

func (m MovieDB) SaveFilmToHat(tgUser int64, film string) error {
	user, err := m.GetOrCreateUser(tgUser)
	if err != nil {
		return err
	}
	movie, err := m.GetOrCreateMovie(film)
	if err != nil {
		return err
	}
	movie.UserID = user.ID
	result := m.db.Save(movie)
	err = result.Error
	if err != nil {
		return err
	}
	return nil
}

func (m MovieDB) GetFilms(status MovieStatus) (Movies, error) {
	movies := Movies{}
	result := m.db.Where("status = ?", status).Find(&movies)
	if result.Error != nil {
		return movies, nil
	}
	return movies, nil
}

func (m MovieDB) DeleteFilmFromHat(movie string) error {
	movieObject, err := m.GetOrCreateMovie(movie)
	if err != nil {
		return err
	}
	movieObject.Status = MovieStatusWatched
	result := m.db.Save(movieObject)
	err = result.Error
	if err != nil {
		return err
	}
	return nil
}
