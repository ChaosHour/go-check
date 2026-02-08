package checker

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Config holds the configuration for the checker
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	Threads  int
	Duration int
	Query    string // Optional query to run
	Interval int    // Interval to re-execute query (0 = once)
}

// ReadMyCnf reads the ~/.my.cnf file to extract MySQL credentials
func ReadMyCnf() (string, string, string, error) {
	homeDir := os.Getenv("HOME")
	cnfFile := homeDir + "/.my.cnf"
	content, err := os.ReadFile(cnfFile)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to read ~/.my.cnf file: %v", err)
	}

	var user, password, host string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "user"):
			user = strings.TrimSpace(strings.SplitN(line, "=", 2)[1])
		case strings.HasPrefix(line, "password"):
			password = strings.TrimSpace(strings.SplitN(line, "=", 2)[1])
		case strings.HasPrefix(line, "host"):
			host = strings.TrimSpace(strings.SplitN(line, "=", 2)[1])
		}
	}

	if user == "" || password == "" {
		return "", "", "", fmt.Errorf("incomplete credentials in ~/.my.cnf file")
	}

	return user, password, host, nil
}

// RunChecker runs the concurrent connection test
func RunChecker(config Config) {
	// Ensure required parameters are set
	if config.User == "" || config.Password == "" || config.Host == "" {
		log.Fatal("MySQL credentials (user, password, host) must be provided")
	}

	// Create a wait group to wait for all goroutines
	var wg sync.WaitGroup

	// MySQL connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=30s&readTimeout=30s&writeTimeout=30s", config.User, config.Password, config.Host, config.Port, config.Database)

	// Start the connections concurrently
	for i := 0; i < config.Threads; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			testMySQLConnection(id, dsn, config.Duration, config.Query, config.Interval)
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	fmt.Printf("All %d connections closed.\n", config.Threads)
}

func testMySQLConnection(id int, dsn string, duration int, query string, interval int) {
	// Open a MySQL connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Goroutine %d: Failed to connect to MySQL: %v\n", id, err)
		return
	}
	defer db.Close()

	// Ping the database to check if the connection is alive
	if err := db.Ping(); err != nil {
		log.Printf("Goroutine %d: Connection error: %v\n", id, err)
		return
	}

	fmt.Printf("Goroutine %d: Connected to MySQL.\n", id)

	if query != "" {
		if interval > 0 {
			// Execute query periodically
			ticker := time.NewTicker(time.Duration(interval) * time.Second)
			defer ticker.Stop()
			endTime := time.Now().Add(time.Duration(duration) * time.Second)

			for {
				select {
				case <-ticker.C:
					start := time.Now()
					rows, err := db.Query(query)
					if err != nil {
						log.Printf("Goroutine %d: Query error: %v\n", id, err)
						return
					}
					rows.Close()
					elapsed := time.Since(start)
					fmt.Printf("Goroutine %d: Query executed in %v\n", id, elapsed)
				default:
					if time.Now().After(endTime) {
						goto done
					}
					time.Sleep(100 * time.Millisecond) // Small sleep to avoid busy waiting
				}
			}
		done:
		} else {
			// Execute query once
			start := time.Now()
			rows, err := db.Query(query)
			if err != nil {
				log.Printf("Goroutine %d: Query error: %v\n", id, err)
				return
			}
			defer rows.Close()
			elapsed := time.Since(start)
			fmt.Printf("Goroutine %d: Query executed in %v\n", id, elapsed)
		}
	} else {
		// Simulate keeping the connection alive for the given duration
		fmt.Printf("Goroutine %d: Keeping connection alive for %d seconds...\n", id, duration)
		time.Sleep(time.Duration(duration) * time.Second)
	}

	fmt.Printf("Goroutine %d: Closing connection.\n", id)
}
