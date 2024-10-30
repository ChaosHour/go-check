package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/time/rate"
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

// ConnectionStats holds metrics about the connection test
type ConnectionStats struct {
	activeConnections int64
	successfulConns   int64
	failedConns       int64
	totalLatency      time.Duration
}

func (s *ConnectionStats) incrementActive() { atomic.AddInt64(&s.activeConnections, 1) }
func (s *ConnectionStats) decrementActive() { atomic.AddInt64(&s.activeConnections, -1) }
func (s *ConnectionStats) addSuccess()      { atomic.AddInt64(&s.successfulConns, 1) }
func (s *ConnectionStats) addFailure()      { atomic.AddInt64(&s.failedConns, 1) }

// Add new ConnectionConfig type
type ConnectionConfig struct {
	batchSize    int
	batchDelay   time.Duration
	rateLimit    float64
	warmupPeriod time.Duration
}

func main() {
	// Add program usage information
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "go-check - MySQL connection testing utility\n\n")
		fmt.Fprintf(os.Stderr, "Usage: go-check [options]\n\nOptions:\n")
		flag.PrintDefaults()
	}

	// CLI arguments
	host := flag.String("host", "", "MySQL server host (overrides ~/.my.cnf if set)")
	port := flag.String("port", "3306", "MySQL server port")
	user := flag.String("user", "", "MySQL username (overrides ~/.my.cnf if set)")
	password := flag.String("password", "", "MySQL password (overrides ~/.my.cnf if set)")
	database := flag.String("database", "test", "MySQL database name")
	threads := flag.Int("threads", 10, "Number of concurrent connections")
	duration := flag.Int("duration", 10, "Duration to keep connections alive (in seconds)")
	maxIdleConns := flag.Int("max-idle-conns", 10, "Maximum number of idle connections")
	maxOpenConns := flag.Int("max-open-conns", 100, "Maximum number of open connections")
	connMaxLifetime := flag.Duration("conn-max-lifetime", 5*time.Minute, "Maximum connection lifetime")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	batchSize := flag.Int("batch-size", 10, "Number of connections to establish in each batch")
	batchDelay := flag.Duration("batch-delay", 500*time.Millisecond, "Delay between connection batches")
	rateLimit := flag.Float64("rate-limit", 0, "Maximum new connections per second (0 = unlimited)")
	warmupPeriod := flag.Duration("warmup", 5*time.Second, "Gradual warmup period")

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

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received shutdown signal. Closing connections...")
		cancel()
	}()

	// Initialize database pool
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=30s&readTimeout=30s&writeTimeout=30s", *user, *password, *host, *port, *database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to create database pool: %v", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxIdleConns(*maxIdleConns)
	db.SetMaxOpenConns(*maxOpenConns)
	db.SetConnMaxLifetime(*connMaxLifetime)

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("go-check: Failed to connect to database: %v", err)
	}

	// Create a wait group to wait for all goroutines
	var wg sync.WaitGroup
	errChan := make(chan error, *threads)

	stats := &ConnectionStats{}

	config := ConnectionConfig{
		batchSize:    *batchSize,
		batchDelay:   *batchDelay,
		rateLimit:    *rateLimit,
		warmupPeriod: *warmupPeriod,
	}

	// Create rate limiter if specified
	var limiter *rate.Limiter
	if config.rateLimit > 0 {
		limiter = rate.NewLimiter(rate.Limit(config.rateLimit), *batchSize)
	}

	// Start stats printer
	if *verbose {
		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					log.Printf("Active: %d, Success: %d, Failed: %d\n",
						atomic.LoadInt64(&stats.activeConnections),
						atomic.LoadInt64(&stats.successfulConns),
						atomic.LoadInt64(&stats.failedConns))
				}
			}
		}()
	}

	// Start connections in batches
	for i := 0; i < *threads; i += config.batchSize {
		if limiter != nil {
			err := limiter.Wait(ctx)
			if err != nil {
				log.Printf("Rate limiter interrupted: %v", err)
				break
			}
		}

		batchThreads := config.batchSize
		if remaining := *threads - i; remaining < batchThreads {
			batchThreads = remaining
		}

		// Start batch of connections
		for j := 0; j < batchThreads; j++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				stats.incrementActive()
				defer stats.decrementActive()

				// Add warmup delay based on connection ID
				if config.warmupPeriod > 0 {
					warmupDelay := time.Duration(float64(config.warmupPeriod) * (float64(id) / float64(*threads)))
					time.Sleep(warmupDelay)
				}

				start := time.Now()
				if err := testMySQLConnection(ctx, db, id, *duration); err != nil {
					stats.addFailure()
					errChan <- fmt.Errorf("goroutine %d: %v", id, err)
				} else {
					stats.addSuccess()
					atomic.AddInt64((*int64)(&stats.totalLatency), int64(time.Since(start)))
				}
			}(i + j)
		}

		// Delay between batches
		time.Sleep(config.batchDelay)
	}

	// Wait for completion or cancellation
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Handle errors and completion
	for err := range errChan {
		log.Printf("Error: %v", err)
	}

	// Print final statistics
	successfulConns := atomic.LoadInt64(&stats.successfulConns)
	failedConns := atomic.LoadInt64(&stats.failedConns)
	totalConns := successfulConns + failedConns
	avgLatency := time.Duration(0)
	if successfulConns > 0 {
		avgLatency = time.Duration(atomic.LoadInt64((*int64)(&stats.totalLatency))) / time.Duration(successfulConns)
	}

	fmt.Printf("\ngo-check test completed:\n")
	fmt.Printf("Total connections attempted: %d\n", totalConns)
	fmt.Printf("Successful connections: %d\n", successfulConns)
	fmt.Printf("Failed connections: %d\n", failedConns)
	fmt.Printf("Success rate: %.2f%%\n", float64(successfulConns)/float64(totalConns)*100)
	fmt.Printf("Average connection latency: %v\n", avgLatency)

	fmt.Printf("All %d connections closed.\n", *threads)
}

func testMySQLConnection(ctx context.Context, db *sql.DB, id int, duration int) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	deadline := time.Now().Add(time.Duration(duration) * time.Second)
	fmt.Printf("Goroutine %d: Connected to MySQL. Keeping connection alive until %v...\n", id, deadline.Format(time.RFC3339))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return nil
			}
			start := time.Now()
			if err := db.PingContext(ctx); err != nil {
				return fmt.Errorf("ping failed (after %v): %v", time.Since(start), err)
			}
		}
	}
}
