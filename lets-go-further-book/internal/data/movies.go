package data

import "time"

type Movie struct {
	ID        int64     `json:"id"`                // Unique identifier for the movie
	CreatedAt time.Time `json:"-"`                 // Time when the movie was added to our db
	Title     string    `json:"title"`             // The title of the movie
	Year      int32     `json:"year,omitempty"`    // The release year of the movie
	Runtime   Runtime   `json:"runtime,omitempty"` // The runtime of the movie in minutes
	Genres    []string  `json:"genres,omitempty"`  // The genres of the movie
	Version   int32     `json:"version"`           // The version of the movie: starts at 1 and increments each time the movie is updated
}
