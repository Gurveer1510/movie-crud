package types

type IMDBResp struct {
	Title      string `json:"Title"`
	Released   string `json:"Released"`
	IMDBRating string `json:"imdbRating"`
	Runtime    string `json:"Runtime"`
	Plot       string `json:"Plot"`
	Response   string `json:"Response"`
}

type APIResponse struct {
	Message string                 `json:"message"`
	Status  string                 `json:"status"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type Movie struct {
	ID             int    `json:"id"`
	MovieName      string `json:"movie_name"`
	Synopsis       string `json:"synopsis"`
	RuntimeMinutes int    `json:"runtime_minutes"`
	IMDBRating     string `json:"imdb_rating"`
	ReleaseDate    string `json:"relase_date"`
}

type CreateMovieReuest struct {
	MovieName string `json:"movie_name" validate:"required"`
}

type UpdateMovieRequest struct {
	MovieName      string `json:"movie_name,omitempty"`
	Synopsis       string `json:"synopsis,omitempty"`
	RuntimeMinutes int    `json:"runtime_minutes,omitempty"`
	IMDBRating     string `json:"imdb_rating,omitempty"`
}
