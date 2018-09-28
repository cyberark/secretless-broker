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

func getConnection(endpoint Endpoint) (*sql.DB, error) {
	var ok bool
	var envAddress string
	if envAddress, ok = endpointToEnv[endpoint]; ok == false {
		return nil, fmt.Errorf("got unknown endpoint %v", endpoint)
	}

	var address string
	if address, ok = os.LookupEnv(envAddress); ok == false {
		return nil, fmt.Errorf("%s is not set", envAddress)
	}

	if endpoint == Postgres {
		address = fmt.Sprintf("test@%s", address)
	}

	connStr := fmt.Sprintf("postgresql://%s/postgres?sslmode=disable", address)
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

func benchmarkQuery(endpoint Endpoint, query string, b *testing.B) {
	b.StopTimer()
	db, err := getConnection(endpoint)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	for i := 0; i < b.N; i++ {
		runQuery(db, query, b)
	}
}

func BenchmarkBaseline_Select1(b *testing.B) {
	benchmarkQuery(Postgres, select1Query, b)
}

func BenchmarkBaseline_Select10(b *testing.B) {
	benchmarkQuery(Postgres, select10Query, b)
}

func BenchmarkBaseline_Select100(b *testing.B) {
	benchmarkQuery(Postgres, select100Query, b)
}

func BenchmarkBaseline_Select1000(b *testing.B) {
	benchmarkQuery(Postgres, select1000Query, b)
}

func BenchmarkBaseline_Select10000(b *testing.B) {
	benchmarkQuery(Postgres, select10000Query, b)
}

func BenchmarkSecretless_Select1(b *testing.B) {
	benchmarkQuery(Secretless, select1Query, b)
}

func BenchmarkSecretless_Select10(b *testing.B) {
	benchmarkQuery(Secretless, select10Query, b)
}

func BenchmarkSecretless_Select100(b *testing.B) {
	benchmarkQuery(Secretless, select100Query, b)
}

func BenchmarkSecretless_Select1000(b *testing.B) {
	benchmarkQuery(Secretless, select1000Query, b)
}

func BenchmarkSecretless_Select10000(b *testing.B) {
	benchmarkQuery(Secretless, select10000Query, b)
}
