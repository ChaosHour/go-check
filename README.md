# go-check

`go-check` is a lightweight CLI tool built in Go to test and benchmark concurrent MySQL connections. It allows you to specify the number of threads (goroutines) that connect to a MySQL server and hold the connections open for a specified duration, simulating concurrent load on the database. This is useful for testing the performance, stability, and connection handling capacity of MySQL instances.

## Features

- Connects to a MySQL server using multiple concurrent connections (threads).
- Allows you to configure the number of threads and the duration for which connections should be kept alive.
- Supports reading MySQL credentials from the `~/.my.cnf` file, simplifying authentication.
- Logs connection lifecycle, indicating when connections are established and closed.

## Use Cases

- **Stress Testing**: Evaluate the MySQL server's ability to handle concurrent connections under different loads.
- **Performance Benchmarking**: Identify connection bottlenecks and timeouts when multiple clients are connected.
- **Testing Connection Limits**: Experiment with thread counts to discover the maximum number of connections your MySQL instance can handle reliably.

## Installation

First, clone the repository and navigate to the project directory:

```bash
git clone https://github.com/ChaosHour/go-check.git
cd go-check
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

```Go
./go-check -host=192.x.x.x -database=sbtest -threads=10 -duration=60

Goroutine 2: Connected to MySQL. Keeping connection alive for 60 seconds...
Goroutine 8: Connected to MySQL. Keeping connection alive for 60 seconds...
Goroutine 6: Connected to MySQL. Keeping connection alive for 60 seconds...
Goroutine 9: Connected to MySQL. Keeping connection alive for 60 seconds...
Goroutine 0: Connected to MySQL. Keeping connection alive for 60 seconds...
Goroutine 4: Connected to MySQL. Keeping connection alive for 60 seconds...
Goroutine 3: Connected to MySQL. Keeping connection alive for 60 seconds...
Goroutine 1: Connected to MySQL. Keeping connection alive for 60 seconds...
Goroutine 5: Connected to MySQL. Keeping connection alive for 60 seconds...
Goroutine 7: Connected to MySQL. Keeping connection alive for 60 seconds...
Goroutine 2: Closing connection.
Goroutine 9: Closing connection.
Goroutine 8: Closing connection.
Goroutine 6: Closing connection.
Goroutine 0: Closing connection.
Goroutine 3: Closing connection.
Goroutine 4: Closing connection.
Goroutine 1: Closing connection.
Goroutine 5: Closing connection.
Goroutine 7: Closing connection.
All 10 connections closed.

```

## To Build

```Go
To build:

go build -o go-gtids

FreeBSD:
env GOOS=freebsd GOARCH=amd64 go build .

On Mac:
env GOOS=darwin GOARCH=amd64 go build .

Linux:
env GOOS=linux GOARCH=amd64 go build .
```
