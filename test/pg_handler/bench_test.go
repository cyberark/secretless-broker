package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	select1Query     = "SELECT 1"
	select10Query    = "SELECT generate_series(0, 9)"
	select100Query   = "SELECT generate_series(0, 99)"
	select1000Query  = "SELECT generate_series(0, 999)"
	select10000Query = "SELECT generate_series(0, 9999)"
)

type Endpoint int

const (
	Postgres Endpoint = iota
	Secretless
)

var endpointToEnv = map[Endpoint]string{
	Postgres:   "PG_ADDRESS",
	Secretless: "SECRETLESS_ADDRESS",
}

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
	b.StartTimer()
	rows, err := db.Query(query)
	if err != nil {
		b.Fatal(err)
	}
	b.StopTimer()
	rows.Close()
}

func benchmarkQuery(query string, b *testing.B) {
	b.StopTimer()
	db, err := getConnection()
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	for i := 0; i < b.N; i++ {
		runQuery(db, query, b)
	}
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
