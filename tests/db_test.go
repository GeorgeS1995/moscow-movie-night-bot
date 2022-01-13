package tests

import (
	"testing"

	"github.com/GeorgeS1995/moscow-movie-night-bot/internal/db"
)

func TestGetMoviewList(t *testing.T) {
	movieList1 := db.Movies{db.Movie{Label: "A"}}
	movieList2 := db.Movies{db.Movie{Label: "A"}, db.Movie{Label: "B"}, db.Movie{Label: "C"}}
	testCases := []db.Movies{movieList1, movieList2}
	for _, tc := range testCases {
		result := tc.GetMoviewList()
		if len(result) != len(tc)*2 {
			t.Error("Wrong movies len")
			continue
		}
	}
	t.Log("TestGetMoviewList OK")
}
