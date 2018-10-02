package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	select1Query     = "SELECT * FROM test.test WHERE id < 1"
	select10Query    = "SELECT * FROM test.test WHERE id < 10"
	select100Query   = "SELECT * FROM test.test WHERE id < 100"
	select1000Query  = "SELECT * FROM test.test WHERE id < 1000"
	select10000Query = "SELECT * FROM test.test WHERE id < 10000"
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

func benchmarkQuery(query string, b *testing.B) {

	// stop time while the connection is opened
	b.StopTimer()

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
	benchmarkQuery(select1Query, b)
}

func Benchmark_Select10(b *testing.B) {
	benchmarkQuery(select10Query, b)
}

func Benchmark_Select100(b *testing.B) {
	benchmarkQuery(select100Query, b)
}

func Benchmark_Select1000(b *testing.B) {
	benchmarkQuery(select1000Query, b)
}

func Benchmark_Select10000(b *testing.B) {
	benchmarkQuery(select10000Query, b)
}
