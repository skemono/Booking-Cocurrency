package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // Import the pq driver for PostgreSQL
	"github.com/pressly/goose/v3"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "admin"
	password = "reservasDB123!"
	dbname   = "reservas_db"
)

// postgresql://admin:reservasDB123!@localhost:5432/reservas_db
func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	// poner dialecto de goose a postgres
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set dialect: %v", err)
	}

	// Migrar la base de datos
	if err := goose.Up(db, "./db"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	fmt.Println("Successfully connected!")
}
