package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type Record struct {
	Address string
}

func main() {
	// Connect to the PostgreSQL database
	db, err := sql.Open("postgres", "user=postgres dbname=Bitcoin sslmode=disable password=admin host=localhost port=5433")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Open the TSV file
	file, err := os.Open("bitcoin_addresses_latest.tsv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a new CSV reader with the tab delimiter for TSV
	reader := csv.NewReader(file)
	reader.Comma = '\t' // Set delimiter to tab for TSV

	// Read all records from the file
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// Prepare the SQL insert statement
	stmt, err := db.Prepare("INSERT INTO rich (address) VALUES ($1)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Process each record and insert into the database
	for _, record := range records {
		address := record[0]

		// Execute the insert statement for each address
		_, err := stmt.Exec(address)
		if err != nil {
			log.Printf("Error inserting address %s: %v\n", address, err)
			continue
		}
	}

	// Confirm the operation
	fmt.Println("Data inserted successfully!")
}
