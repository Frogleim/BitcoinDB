package main

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	_ "github.com/lib/pq"
)

// Database connection details
const (
	dbHost     = "localhost"
	dbPort     = 5433
	dbUser     = "postgres"
	dbPassword = "admin"
	dbName     = "Bitcoin"
)

// GenerateBitcoinKey generates a private key, public key, WIF, and address
func GenerateBitcoinKey() (string, string, string, string, error) {
	// Generate a new private key using secp256k1
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		return "", "", "", "", fmt.Errorf("error generating private key: %w", err)
	}

	// Convert the private key to hexadecimal
	privateKeyHex := hex.EncodeToString(privateKey.Serialize())

	// Convert the private key to WIF
	wif, err := btcutil.NewWIF(privateKey, &chaincfg.MainNetParams, true)
	if err != nil {
		return "", "", "", "", fmt.Errorf("error generating WIF: %w", err)
	}

	// Generate the public key
	publicKey := privateKey.PubKey()
	publicKeyHex := hex.EncodeToString(publicKey.SerializeCompressed())

	// Generate the Bitcoin address
	address, err := btcutil.NewAddressPubKey(publicKey.SerializeCompressed(), &chaincfg.MainNetParams)
	if err != nil {
		return "", "", "", "", fmt.Errorf("error generating address: %w", err)
	}

	return privateKeyHex, publicKeyHex, wif.String(), address.EncodeAddress(), nil
}

// InsertIntoDB inserts the generated Bitcoin key details into the database
func InsertIntoDB(db *sql.DB, address, publicKey, wif, privateKey string) error {
	query := `
		INSERT INTO bitcoin_keys (address, public_key, wif, private_key)
		VALUES ($1, $2, $3, $4)
	`

	_, err := db.Exec(query, address, publicKey, wif, privateKey)
	if err != nil {
		return fmt.Errorf("error inserting into database: %w", err)
	}

	return nil
}

func main() {
	// Connect to the database
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

	// Infinite loop to generate and insert Bitcoin keys into the database
	for {
		// Generate Bitcoin keys and save to the database
		privateKey, publicKey, wif, address, err := GenerateBitcoinKey()
		if err != nil {
			log.Printf("Error generating Bitcoin key: %v\n", err)
			continue
		}

		fmt.Printf("Generated Bitcoin Key:\n")
		fmt.Printf("  Address: %s\n", address)
		fmt.Printf("  Public Key: %s\n", publicKey)
		fmt.Printf("  WIF: %s\n", wif)
		fmt.Printf("  Private Key: %s\n\n", privateKey)

		// Insert into the database
		err = InsertIntoDB(db, address, publicKey, wif, privateKey)
		if err != nil {
			log.Printf("Error inserting into database: %v\n", err)
			continue
		}

		log.Printf("Inserted Bitcoin key for address %s into the database.\n", address)

		// Delay between iterations to avoid too many requests
	}
}
