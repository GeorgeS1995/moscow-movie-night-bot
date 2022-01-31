package db

import (
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
