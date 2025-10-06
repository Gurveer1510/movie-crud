package main

import (
	"context"
	"fmt"
	"log"
	"movie-crud-api/db"
	"movie-crud-api/handlers"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main () {
	err := godotenv.Load()
	if err != nil {
		log.Println("No ENV File found.")
	}
	db.ConnectDb()
	defer db.Conn.Close(context.Background())
	
	http.HandleFunc("/movies/create", handlers.CreateMovie)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running at http://localhost:%s", port)
	http.ListenAndServe(fmt.Sprintf(":%v",port), nil)
}