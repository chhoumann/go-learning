package data

import (
	"time"

	"greenlight.bagerbach.com/internal/validator"
)

type Movie struct {
	ID        int64     `json:"id"`                // Unique identifier for the movie
	CreatedAt time.Time `json:"-"`                 // Time when the movie was added to our db
	Title     string    `json:"title"`             // The title of the movie
	Year      int32     `json:"year,omitempty"`    // The release year of the movie
	Runtime   Runtime   `json:"runtime,omitempty"` // The runtime of the movie in minutes
	Genres    []string  `json:"genres,omitempty"`  // The genres of the movie
	Version   int32     `json:"version"`           // The version of the movie: starts at 1 and increments each time the movie is updated
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}
