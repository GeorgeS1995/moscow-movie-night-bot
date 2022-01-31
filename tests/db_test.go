package tests

import (
	"testing"

	"github.com/GeorgeS1995/moscow-movie-night-bot/internal/db"
)

func TestGetMoviesList(t *testing.T) {
	movieList1 := db.Movies{db.Movie{Label: "A"}}
	movieList2 := db.Movies{db.Movie{Label: "A"}, db.Movie{Label: "B"}, db.Movie{Label: "C"}}
	testCases := []db.Movies{movieList1, movieList2}
	for idx, tc := range testCases {
		for _, seq := range []bool{true} {
			copiedTC := make(db.Movies, len(tc))
			copy(copiedTC, tc)
			result := tc.GetMoviesList(seq)
			if len(result) != len(tc)*2 {
				t.Error("Wrong movies len")
				continue
			}
			// Check ordered seq
			if idx > 0 && !seq && result != "A\nB\nC\n" {
				t.Errorf("Wrong ordered sequincies, expected: ABC, got: %s", result)
			}
		}
	}
	t.Log("TestGetMoviesList OK")
}
