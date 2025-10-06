package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

var Conn *pgx.Conn

func ConnectDb() {
	
	database_url := os.Getenv("DATABASE_URL")
	if database_url == "" {
		log.Fatal("DATABASE_URL not set in .env")
	}
	var err error
	Conn, err = pgx.Connect(context.Background(), database_url)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n",err)
	}
	
	log.Println("Connected to NeonDB")
	queryBytes, err := os.ReadFile("db/queries/create_table.sql")
	if err != nil {
		log.Fatalf("Failed to read SQL File: %v\n", err)
	}

	query := string(queryBytes)

	_, err = Conn.Exec(context.Background(), query)
	if err != nil {
		log.Fatalf("Failed to run migration: %v\n", err)
	}

	log.Println("Movie Table is READY!! ^_^")
	
}