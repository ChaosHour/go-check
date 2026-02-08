package main

import (
	"flag"
	"log"

	"github.com/ChaosHour/go-check/pkg/checker"
)

func main() {
	// CLI arguments
	host := flag.String("host", "", "MySQL server host (overrides ~/.my.cnf if set)")
	port := flag.String("port", "3306", "MySQL server port")
	user := flag.String("user", "", "MySQL username (overrides ~/.my.cnf if set)")
	password := flag.String("password", "", "MySQL password (overrides ~/.my.cnf if set)")
	database := flag.String("database", "test", "MySQL database name")
	threads := flag.Int("threads", 10, "Number of concurrent connections")
	duration := flag.Int("duration", 10, "Duration to keep connections alive (in seconds)")
	query := flag.String("query", "", "Optional SQL query to run on each connection")
	interval := flag.Int("interval", 0, "Interval to re-execute query (in seconds, 0 = execute once)")

	flag.Parse()

	config := checker.Config{
		Host:     *host,
		Port:     *port,
		User:     *user,
		Password: *password,
		Database: *database,
		Threads:  *threads,
		Duration: *duration,
		Query:    *query,
		Interval: *interval,
	}

	// Attempt to read credentials from ~/.my.cnf if user or password is not provided
	if config.User == "" || config.Password == "" {
		cnfUser, cnfPassword, _, err := checker.ReadMyCnf()
		if err != nil {
			log.Fatalf("Error reading credentials from ~/.my.cnf: %v", err)
		}

		// Use ~/.my.cnf values if CLI arguments are not provided
		if config.User == "" {
			config.User = cnfUser
		}
		if config.Password == "" {
			config.Password = cnfPassword
		}
	}

	checker.RunChecker(config)
}
