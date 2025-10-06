package models

import "time"

type Movie struct {
	ID      int `json:"id"`
	MovieName string `json:"movie_name"`
	Synopsis string `json:"synopsis"`
	RuntimeMinutes int `json:"runtime_minutes"`
	IMDBRating string `json:"imdb_rating"`
	ReleaseDate *time.Time `json:"relase_date"`
}

type CreateMovieReuest struct {
    MovieName string `json:"movie_name" validate:"required"`
    Synopsis string `json:"synopsis" validate:"required"`
    RuntimeMinutes int `json:"runtime_minutes" validate:"required"`
    IMDBRating string `json:"imdb_rating"`
    ReleaseDate string `json:"relase_date"`
}

