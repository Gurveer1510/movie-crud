package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"movie-crud-api/db"
	"movie-crud-api/logger"
	"movie-crud-api/types"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func CreateMovie(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	var input types.CreateMovieReuest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid json payload", http.StatusBadRequest)
		return
	}

	if len(strings.TrimSpace(input.MovieName)) == 0 {
		http.Error(w, "Movie Name is required", http.StatusBadRequest)
		return
	}

	omdbAPI := os.Getenv("API_KEY")
	if omdbAPI == "" {
		http.Error(w, "API KEY NOT FOUND FOR OMDB", http.StatusInternalServerError)
		return
	}
	baseUrl := "https://www.omdbapi.com/"
	movieName := input.MovieName
	params := url.Values{}
	params.Add("t", movieName)
	params.Add("apikey", omdbAPI)
	imdbUrl := fmt.Sprintf("%s?%s", baseUrl, params.Encode())
	response, err := http.Get(imdbUrl)

	if err != nil {
		http.Error(w, fmt.Sprintf("No Movie found for the name: %v", input.MovieName), http.StatusNotFound)
	}

	defer response.Body.Close()

	resBody, respErr := io.ReadAll(response.Body)
	if respErr != nil {
		fmt.Println("Error reading Body:", respErr)
		return
	}

	var result types.IMDBResp

	parsingError := json.Unmarshal(resBody, &result)

	if result.Response == "False" {
		http.Error(w, fmt.Sprintf("No Movie found for the name: %v", input.MovieName), http.StatusNotFound)
	}

	if parsingError != nil {
		fmt.Println("Error decoding imdb data:", parsingError)
		return
	}
	fmt.Printf("Title: %s\nReleased: %s\nRating: %s\nRuntime: %s\nPlot: %s",
		result.Title, result.Released, result.IMDBRating, result.Runtime, result.Plot)

	var formatted string
	t, parseError := time.Parse("02 Jan 2006", result.Released)
	if parseError != nil {
		fmt.Println("Found Incorrect format for the release date\n", parseError)
		formatted = "N/A"
	} else {
		formatted = t.Format("2006-01-02")
	}

	runtimeStr := strings.Split(result.Runtime, " ")[0]
	runtime, err := strconv.Atoi(runtimeStr)
	if err != nil {
		fmt.Println("Couldn't get runtime for ", input.MovieName)
		runtime = 0
	}

	movie := types.Movie{
		MovieName:      input.MovieName,
		Synopsis:       result.Plot,
		RuntimeMinutes: runtime,
		IMDBRating:     result.IMDBRating,
		ReleaseDate:    formatted,
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
	err = db.Conn.QueryRow(context.Background(),
		query,
		movie.MovieName,
		result.Plot,
		movie.RuntimeMinutes,
		result.IMDBRating,
		movie.ReleaseDate,
	).Scan(&movie.ID, &createdAt, &updatedAt)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		types.Movie
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}{
		Movie:     movie,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("/movies/%s", url.PathEscape(movie.MovieName)))
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func GetAllMovies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	query := `SELECT id, movie_name, synopsis, runtime_minutes, imdb_rating, release_date FROM movies WHERE is_deleted = FALSE;`
	rows, err := db.Conn.Query(context.Background(), query)
	if err != nil {
		http.Error(w, "Error in getting movies from database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var movies []types.Movie
	for rows.Next() {
		var m types.Movie
		var releaseDate time.Time
		err := rows.Scan(&m.ID, &m.MovieName, &m.Synopsis, &m.RuntimeMinutes, &m.IMDBRating, &releaseDate)
		if err != nil {
			http.Error(w, "Error scanning movie row", http.StatusInternalServerError)
			return
		}
		m.ReleaseDate = releaseDate.Format("2006-01-02")
		movies = append(movies, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

func GetMovieByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}
	reqQuery := r.URL.Query()

	movie_name := reqQuery.Get("movie_name")
	if movie_name == "" {
		http.Error(w, "Bad Request", http.StatusNotFound)
		return
	}

	query := `
		SELECT id, movie_name, synopsis, runtime_minutes, imdb_rating, release_date,
       SIMILARITY(movie_name, $1) AS sim
FROM movies
WHERE movie_name % $2 AND is_deleted = FALSE
ORDER BY sim DESC;
	`
	_, _ = db.Conn.Exec(context.Background(), "SET pg_trgm.similarity_threshold = 0.2")
	rows, err := db.Conn.Query(context.Background(), query, movie_name, movie_name)

	if err != nil {
		http.Error(w, "Error in getting movies from database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var movies []types.Movie
	for rows.Next() {
		var m types.Movie
		var sim float64
		var rd time.Time
		err := rows.Scan(&m.ID, &m.MovieName, &m.Synopsis, &m.RuntimeMinutes, &m.IMDBRating, &rd, &sim)
		m.ReleaseDate = rd.Format("2006-01-02")
		if err != nil {
			newError := types.APIResponse{
				Message: "could not read data from database",
				Status:  "failed",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(newError)
			return
		}
		log.Println(m)

		movies = append(movies, m)
	}

	if len(movies) == 0 {
		newError := types.APIResponse{
			Message: "No movie found",
			Status:  "Failed",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&newError)
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":  len(movies),
		"movies": movies,
	})
}

func UpdateMovieById(w http.ResponseWriter, r *http.Request) {
	logger.Info("Updating Movie by Id")
	if r.Method != http.MethodPut {
		newErr := types.APIResponse{
			Message: "Invalid Method",
			Status:  "failed",
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(newErr)
	}

	updateReq := types.UpdateMovieRequest{}
	err := json.NewDecoder(r.Body).Decode(&updateReq)
	if err != nil {
		logger.Error("DECODING")
		newErr := types.APIResponse{
			Message: "Bad Payload",
			Status:  "Failed",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(newErr)
	}

}
