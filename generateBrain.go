package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg" // Import for MainNetParams
	_ "github.com/lib/pq"               // PostgreSQL driver
)

const (
	dbName     = "Bitcoin"
	dbUser     = "postgres"
	dbPassword = "admin"
	dbHost     = "localhost"
	dbPort     = 5433
)

func main() {
	// Read passphrases from the file
	passphrases, err := readPassphrases("realhuman_phill.txt")
	if err != nil {
		log.Fatalf("Error reading passphrases: %v", err)
	}
	fmt.Printf("Number of passphrases: %d\n", len(passphrases))

	// Open the results file
	outputFile, err := os.Create("results.txt")
	if err != nil {
		log.Fatalf("Error creating results file: %v", err)
	}
	defer outputFile.Close()

	// Connect to the database
	db, err := connectToDB()
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer db.Close()

	// Process each passphrase
	for _, passphrase := range passphrases {
		privateKey, wif, publicKey, address, err := generateWallet(passphrase)
		if err != nil {
			log.Printf("Error generating wallet for passphrase '%s': %v", passphrase, err)
			continue
		}

		// Prepare wallet data
		walletData := map[string]string{
			"passphrase":  passphrase,
			"private_key": privateKey,
			"wif":         wif,
			"public_key":  publicKey,
			"address":     address,
		}

		// Write wallet data to the results file
		err = writeToFile(outputFile, walletData)
		if err != nil {
			log.Printf("Error writing wallet data to file: %v", err)
		}

		// Insert wallet data into the database
		err = insertWallet(db, walletData)
		if err != nil {
			log.Printf("Error inserting wallet data into database: %v", err)
		} else {
			fmt.Println("Wallet data inserted successfully!")
		}
	}
}

// readPassphrases reads passphrases from a file and returns them as a slice
func readPassphrases(filename string) ([]string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return splitLines(string(data)), nil
}

// splitLines splits the file content into individual lines
func splitLines(data string) []string {
	lines := strings.Split(data, "\n")
	result := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

// connectToDB connects to the PostgreSQL database
func connectToDB() (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"dbname=%s user=%s password=%s host=%s port=%d sslmode=disable",
		dbName, dbUser, dbPassword, dbHost, dbPort,
	)
	return sql.Open("postgres", connStr)
}

// generateWallet generates wallet data from a passphrase
func generateWallet(passphrase string) (string, string, string, string, error) {
	// Generate private key from passphrase
	hash := sha256.Sum256([]byte(passphrase))
	privateKeyBytes := hash[:]

	// Generate private key and public key
	privKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)
	pubKey := privKey.PubKey()

	// Convert private key to WIF
	wif, err := btcutil.NewWIF(privKey, &chaincfg.MainNetParams, true)
	if err != nil {
		return "", "", "", "", err
	}

	// Serialize public key
	publicKey := hex.EncodeToString(pubKey.SerializeCompressed())

	// Generate Bitcoin address
	address, err := btcutil.NewAddressPubKey(pubKey.SerializeCompressed(), &chaincfg.MainNetParams)
	if err != nil {
		return "", "", "", "", err
	}

	return hex.EncodeToString(privKey.Serialize()), wif.String(), publicKey, address.EncodeAddress(), nil
}

// writeToFile writes wallet data to the results file
func writeToFile(file *os.File, walletData map[string]string) error {
	line := fmt.Sprintf(
		"Passphrase: %s, Private Key: %s, WIF: %s, Public Key: %s, Address: %s\n",
		walletData["passphrase"], walletData["private_key"], walletData["wif"],
		walletData["public_key"], walletData["address"],
	)
	_, err := file.WriteString(line)
	return err
}

// insertWallet inserts wallet data into the PostgreSQL database
func insertWallet(db *sql.DB, walletData map[string]string) error {
	query := `
		INSERT INTO wallets (passphrase, private_key, wif, public_key, address)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := db.Exec(
		query,
		walletData["passphrase"], walletData["private_key"], walletData["wif"],
		walletData["public_key"], walletData["address"],
	)
	return err
}
