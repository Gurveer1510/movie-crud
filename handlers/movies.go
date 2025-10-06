package handlers

import (
	"context"
	"encoding/json"
	"movie-crud-api/db"
	"movie-crud-api/models"
	"net/http"
	"time"
)

type MovieResp struct {
	Released string `json:"Released"`
	IMDBRating string `json:"imdbRating"`
}

func CreateMovie(w http.ResponseWriter, r *http.Request){
	if r.Method !=http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	var input models.CreateMovieReuest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid json payload", http.StatusBadRequest)
		return
	}

	var releaseDate *time.Time
	if input.ReleaseDate != "" {
		t, err := time.Parse("2006-01-02", input.ReleaseDate)
		if err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		releaseDate = &t
	}

	movie := models.Movie{
		MovieName: input.MovieName,
		Synopsis: input.Synopsis,
		RuntimeMinutes: input.RuntimeMinutes,
		IMDBRating: input.IMDBRating,
		ReleaseDate: releaseDate,
	}

	if db.Conn == nil {
		http.Error(w, "DB not connected", http.StatusInternalServerError)
		return
	}

	query := `
	INSERT INTO movies (movie_name, synopsis, runtime_minutes, imdb_rating, release_date)
	VALUES
	($1,$2,$3,$4,$5)
	RETURNING id, created_at, updated_at;
	`
	
	var createdAt, updatedAt time.Time
	err := db.Conn.QueryRow(context.Background(), 
	query,
	movie.MovieName,
	movie.Synopsis,
	movie.RuntimeMinutes,
	movie.IMDBRating,
	movie.ReleaseDate,
	).Scan(&movie.ID, &createdAt, &updatedAt)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		models.Movie
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}{
		Movie: movie,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}