package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// readMyCnf reads the ~/.my.cnf file to extract MySQL credentials
func readMyCnf() (string, string, string, error) {
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

	if user == "" || password == "" || host == "" {
		return "", "", "", fmt.Errorf("incomplete credentials in ~/.my.cnf file")
	}

	return user, password, host, nil
}

func main() {
	// CLI arguments
	host := flag.String("host", "", "MySQL server host (overrides ~/.my.cnf if set)")
	port := flag.String("port", "3306", "MySQL server port")
	user := flag.String("user", "", "MySQL username (overrides ~/.my.cnf if set)")
	password := flag.String("password", "", "MySQL password (overrides ~/.my.cnf if set)")
	database := flag.String("database", "test", "MySQL database name")
	threads := flag.Int("threads", 10, "Number of concurrent connections")
	duration := flag.Int("duration", 10, "Duration to keep connections alive (in seconds)")

	flag.Parse()

	// Attempt to read credentials from ~/.my.cnf if user or password is not provided
	if *user == "" || *password == "" || *host == "" {
		cnfUser, cnfPassword, cnfHost, err := readMyCnf()
		if err != nil {
			log.Fatalf("Error reading credentials from ~/.my.cnf: %v", err)
		}

		// Use ~/.my.cnf values if CLI arguments are not provided
		if *user == "" {
			*user = cnfUser
		}
		if *password == "" {
			*password = cnfPassword
		}
		if *host == "" {
			*host = cnfHost
		}
	}

	// Ensure required parameters are set
	if *user == "" || *password == "" || *host == "" {
		log.Fatal("MySQL credentials (user, password, host) must be provided either via CLI or ~/.my.cnf")
	}

	// Create a wait group to wait for all goroutines
	var wg sync.WaitGroup

	// MySQL connection string
	//dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", *user, *password, *host, *port, *database)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=30s&readTimeout=30s&writeTimeout=30s", *user, *password, *host, *port, *database)

	// Start the connections concurrently
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			testMySQLConnection(id, dsn, *duration)
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	fmt.Printf("All %d connections closed.\n", *threads)
}

func testMySQLConnection(id int, dsn string, duration int) {
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

	// Simulate keeping the connection alive for the given duration
	fmt.Printf("Goroutine %d: Connected to MySQL. Keeping connection alive for %d seconds...\n", id, duration)
	time.Sleep(time.Duration(duration) * time.Second)

	fmt.Printf("Goroutine %d: Closing connection.\n", id)
}
