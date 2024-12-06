package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

const (
	dbHost     = "localhost"
	dbPort     = 5433
	dbUser     = "postgres"
	dbPassword = "admin"
	dbName     = "Bitcoin"
)

func main() {
	// Connect to the PostgreSQL database
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v\n", err)
	}
	defer db.Close()

	// Check if the connection is successful
	if err := db.Ping(); err != nil {
		log.Fatalf("Database connection error: %v\n", err)
	}

	// SQL query to get addresses that exist in both tables
	query := `
		SELECT address
		FROM bitcoin_keys
		WHERE address IN (SELECT address FROM rich);
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Error executing query: %v\n", err)
	}
	defer rows.Close()

	// Iterate through the results and print matching addresses
	fmt.Println("Matching addresses:")
	for rows.Next() {
		var address string
		if err := rows.Scan(&address); err != nil {
			log.Fatalf("Error scanning row: %v\n", err)
		}
		fmt.Println(address)
	}

	// Check for any errors encountered during iteration
	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating over rows: %v\n", err)
	}
}
