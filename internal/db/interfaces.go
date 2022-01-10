package db

type MovieDBInterface interface {
	SaveFilmToHat(tgUser int64, film string) error
	GetAllFilms() (Movies, error)
	DeleteFilmFromHat(movie string) error
}
