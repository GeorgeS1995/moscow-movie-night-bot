//go:generate go-enum -f=$GOFILE
package db

// ENUM(Watched, Unwatched)
type MovieStatus int
