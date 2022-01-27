package db

import (
	"errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MovieDB struct {
	db *gorm.DB
}

func OpenDB(dbname string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbname), &gorm.Config{})
	if err != nil {
		return &gorm.DB{}, err
	}
	return db, nil
}

func NewMovieDB(db *gorm.DB) MovieDB {
	return MovieDB{db: db}
}

func MigrateMovieDB(db *gorm.DB) error {
	return db.AutoMigrate(&Movie{}, &User{})
}

func InitMovieDB(dbname string) (MovieDB, error) {
	db, err := OpenDB(dbname)
	if err != nil {
		return MovieDB{}, err
	}
	err = MigrateMovieDB(db)
	if err != nil {
		return MovieDB{}, err
	}
	return NewMovieDB(db), nil
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
		return movies, result.Error
	}
	return movies, nil
}

func (m MovieDB) GetSingleFilm(movieFilter Movie) (SingleMovie, error) {
	movie := SingleMovie{}
	result := m.db.Model(&Movie{}).Select("movies.id, movies.label,movies.status,users.telegram_id").Joins("join users on movies.user_id = users.id").Where(&movieFilter).Find(&movie)
	if result.Error != nil {
		return movie, result.Error
	}
	return movie, nil
}

func (m MovieDB) UpdateSingleFilm(filmID uint, newLabel string) error {
	movie := Movie{}
	result := m.db.First(&movie, filmID)
	if result.Error != nil {
		return result.Error
	}
	movie.Label = newLabel
	result = m.db.Save(&movie)
	if result.Error != nil {
		return result.Error
	}
	return nil
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
