package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func getConnection() (*sql.DB, error) {
	var ok bool
	var address string
	if address, ok = os.LookupEnv("BENCH_ADDRESS"); ok == false {
		return nil, fmt.Errorf("%s is not set", "BENCH_ADDRESS")
	}

	connStr := fmt.Sprintf("postgresql://test@%s/postgres?sslmode=disable", address)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	return db, nil
}

// runQuery executes a query. Expects the timer to already have been stopped.
func runQuery(db *sql.DB, query string, b *testing.B) {
	rows, err := db.Query(query)
	if err != nil {
		b.Fatal(err)
	}
	rows.Close()
}

func benchmarkQuery(queryMax int, b *testing.B) {

	// stop time while the connection is opened
	b.StopTimer()

	query := fmt.Sprintf("SELECT * FROM test.test WHERE id < %d", queryMax)

	db, err := getConnection()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		runQuery(db, query, b)
	}
	b.StopTimer()
}

func Benchmark_Select1(b *testing.B) {
	benchmarkQuery(1, b)
}

func Benchmark_Select10(b *testing.B) {
	benchmarkQuery(10, b)
}

func Benchmark_Select100(b *testing.B) {
	benchmarkQuery(100, b)
}

func Benchmark_Select1000(b *testing.B) {
	benchmarkQuery(1000, b)
}

func Benchmark_Select10000(b *testing.B) {
	benchmarkQuery(10000, b)
}

func Benchmark_Select100000(b *testing.B) {
	benchmarkQuery(100000, b)
}
