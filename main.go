package main

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Found represents the structure of data returned by the query
type Found struct {
	Address       string `db:"address"`
	TotalReceived int64  `db:"total_received"`
	NTx           int64  `db:"n_tx"`
	FinalBalance  int64  `db:"final_balance"`
}

func main() {
	// Connect to a PostgreSQL database
	db, err := sqlx.Connect("postgres", "user=postgres dbname=Bitcoin sslmode=disable password=admin host=localhost port=5433")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	// Test the connection to the database
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfully Connected")
	}

	// Query to fetch data
	rows, err := db.Queryx("SELECT address, total_received, n_tx, final_balance FROM checked_wallets WHERE total_received::BIGINT > 0;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Iterate over rows and scan into the Found struct
	for rows.Next() {
		var place Found
		err := rows.StructScan(&place)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Address: %s, Total Received: %d, Transactions: %d, Final Balance: %d\n",
			place.Address, place.TotalReceived, place.NTx, place.FinalBalance)
	}

	// Check for any errors after iterating over rows
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}
