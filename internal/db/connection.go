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
			movie = Movie{Label: film}
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
	association := m.db.Model(&user).Association("Movies")
	if association.DB.Error != nil {
		return err
	}
	err = association.Append([]Movie{movie})
	if err != nil {
		return err
	}
	return nil
}

func (m MovieDB) GetAllFilms() (Movies, error) {
	movies := Movies{}
	result := m.db.Model(&Movie{}).Select("label").Joins("join user_movies on movies.id=user_movies.movie_id").Scan(&movies)
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
	association := m.db.Model(&movieObject).Association("Users")
	if association.DB.Error != nil {
		return err
	}
	err = association.Clear()
	if association.DB.Error != nil {
		return err
	}
	return nil
}
