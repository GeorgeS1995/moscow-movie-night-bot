package db

type MovieDBInterface interface {
	SaveFilmToHat(tgUser int64, film string) error
	GetFilms(status MovieStatus) (Movies, error)
	DeleteFilmFromHat(movie string) error
	GetSingleFilm(movieFilter Movie) (SingleMovie, error)
	UpdateSingleFilm(filmID uint, newLabel string) error
}
