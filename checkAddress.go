package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
)

// Response struct to parse the JSON response
type Response struct {
	FinalBalance  int64 `json:"final_balance"`
	TotalReceived int64 `json:"total_received"`
	NTx           int64 `json:"n_tx"`
}

// checkBalance fetches balance details from the blockchain API
func checkBalance(address string) (*Response, error) {
	url := fmt.Sprintf("https://blockchain.info/balance?active=%s&base=BTC&cors=true", address)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add headers to the request
	req.Header.Add("accept", "text/html, */*; q=0.01")
	req.Header.Add("accept-language", "en-US,en;q=0.9,ru-RU;q=0.8,ru;q=0.7")
	req.Header.Add("priority", "u=1, i")
	req.Header.Add("sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", `"macOS"`)
	req.Header.Add("sec-fetch-dest", "empty")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("sec-fetch-site", "cross-site")
	req.Header.Add("Referer", "https://privatekeyfinder.io/")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error: received status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Raw JSON Response: %s\n", body)

	var result map[string]Response
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	respData, exists := result[address]
	if !exists {
		return nil, fmt.Errorf("Address %s not found in response", address)
	}

	return &respData, nil
}

func insertIntoDB(db *sql.DB, address string, totalReceived, nTx, finalBalance int64) error {
	query := `
		INSERT INTO checked_wallets (address, total_received, n_tx, final_balance)
		VALUES ($1, $2, $3, $4)
	`

	_, err := db.Exec(query, address, totalReceived, nTx, finalBalance)
	if err != nil {
		return fmt.Errorf("Error inserting data: %w", err)
	}

	log.Printf("Inserted data for address: %s\n", address)
	return nil
}

func main() {
	// Connect to the database
	connStr := "user=postgres dbname=Bitcoin sslmode=disable password=admin host=localhost port=5433"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v\n", err)
	}
	defer db.Close()

	// Check if the connection is successful
	if err := db.Ping(); err != nil {
		log.Fatalf("Database connection error: %v\n", err)
	}

	// Read the file containing Bitcoin addresses
	filePath := "results.txt"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	regexPattern := `Bitcoin Address: ([13][a-km-zA-HJ-NP-Z1-9]{25,34})`
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		log.Fatal(err)
	}

	addressMatches := re.FindAllStringSubmatch(string(content), -1)
	var bitcoinAddresses []string
	for _, match := range addressMatches {
		if len(match) > 1 {
			bitcoinAddresses = append(bitcoinAddresses, match[1])
		}
	}

	fmt.Printf("Number of Bitcoin addresses found: %d\n", len(bitcoinAddresses))

	for _, address := range bitcoinAddresses {
		balance, err := checkBalance(address)
		if err != nil {
			log.Printf("Error checking balance for address %s: %v\n", address, err)
			continue
		}

		//fmt.Printf("Balance for address %s:\n", address)
		//fmt.Printf("  Final Balance: %d\n", balance.FinalBalance)
		//fmt.Printf("  Total Received: %d\n", balance.TotalReceived)
		//fmt.Printf("  Transaction Count: %d\n", balance.NTx)

		// Insert balance details into the database
		err = insertIntoDB(db, address, balance.TotalReceived, balance.NTx, balance.FinalBalance)
		if err != nil {
			log.Printf("Error inserting data for address %s: %v\n", address, err)
		}

		//time.Sleep(100 * time.Millisecond)
	}
}
