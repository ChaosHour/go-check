# go-check

`go-check` is a lightweight CLI tool built in Go to test and benchmark concurrent MySQL connections. It allows you to specify the number of threads (goroutines) that connect to a MySQL server and either hold the connections open for a specified duration or execute a given SQL query concurrently. This is useful for testing the performance, stability, and connection handling capacity of MySQL instances.

## Features

- Connects to a MySQL server using multiple concurrent connections (threads).
- Supports two modes: hold connections alive or execute SQL queries.
- Allows you to configure the number of threads and the duration for which connections should be kept alive.
- Supports reading MySQL credentials from the `~/.my.cnf` file, simplifying authentication.
- Logs connection lifecycle, indicating when connections are established and closed.
- Built with standard Go project structure.

## Use Cases

- **Stress Testing**: Evaluate the MySQL server's ability to handle concurrent connections under different loads.
- **Performance Benchmarking**: Identify connection bottlenecks and timeouts when multiple clients are connected.
- **Testing Connection Limits**: Experiment with thread counts to discover the maximum number of connections your MySQL instance can handle reliably.
- **Query Load Testing**: Run concurrent SELECT queries to simulate read-heavy workloads, e.g., during database upgrades.

## Installation

First, clone the repository and navigate to the project directory:

```bash
git clone https://github.com/ChaosHour/go-check.git
cd go-check
```

Build the binary:

```bash
make build
```

## Credentials from ~/.my.cnf

Using Credentials from ~/.my.cnf
If you have your MySQL credentials stored in a ~/.my.cnf file, go-check can automatically read them. Your ~/.my.cnf file should look something like this:

```bash
[client]
user = your_mysql_user
password = your_mysql_password
host = 127.0.0.1
```

## Running the Tool

### Hold Connections Mode (default)

```bash
./bin/go-check -host=192.x.x.x -database=sbtest -threads=10 -duration=60
```

### Query Execution Mode

```bash
# Execute query once
./bin/go-check -host=192.x.x.x -database=sbtest -threads=10 -query="SELECT COUNT(*) FROM users"

# Execute query every 30 seconds for 5 minutes (300 seconds)
./bin/go-check -host=192.x.x.x -database=sbtest -threads=10 -query="SELECT COUNT(*) FROM users" -interval=30 -duration=300
```

### Example Output

```
Goroutine 2: Connected to MySQL.
Goroutine 8: Connected to MySQL.
Goroutine 6: Connected to MySQL.
Goroutine 9: Connected to MySQL.
Goroutine 0: Connected to MySQL.
Goroutine 4: Connected to MySQL.
Goroutine 3: Connected to MySQL.
Goroutine 1: Connected to MySQL.
Goroutine 5: Connected to MySQL.
Goroutine 7: Connected to MySQL.
Goroutine 2: Query executed in 12.5ms
Goroutine 9: Query executed in 15.2ms
...
Goroutine 2: Closing connection.
Goroutine 9: Closing connection.
...
All 10 connections closed.
```

## CLI Options

- `-host`: MySQL server host (overrides ~/.my.cnf if set)
- `-port`: MySQL server port (default: 3306)
- `-user`: MySQL username (overrides ~/.my.cnf if set)
- `-password`: MySQL password (overrides ~/.my.cnf if set)
- `-database`: MySQL database name (default: test)
- `-threads`: Number of concurrent connections (default: 10)
- `-duration`: Duration to keep connections alive in seconds (default: 10)
- `-query`: Optional SQL query to execute on each connection
- `-interval`: Interval to re-execute query in seconds (default: 0 = execute once, requires -query)

## Project Structure

This project follows the standard Go project layout:

```
go-check/
├── cmd/check/          # Main application entry point
│   └── main.go
├── pkg/checker/        # Core business logic
│   ├── checker.go
│   └── checker_test.go
├── Makefile            # Build and test automation
├── go.mod
├── go.sum
└── README.md
```

## Development

### Build

```bash
make build
```

### Test

```bash
make test
```

### Clean

```bash
make clean
```

### Format Code

```bash
make fmt
```

### Vet Code

```bash
make vet
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
