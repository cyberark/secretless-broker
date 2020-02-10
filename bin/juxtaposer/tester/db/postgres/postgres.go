package postgres

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"github.com/cyberark/secretless-broker/bin/juxtaposer/tester/api"
	sql_db_tester "github.com/cyberark/secretless-broker/bin/juxtaposer/tester/db/sql"
)

type PostgresTester struct {
	sql_db_tester.SqlDatabaseTester
}

func NewTester() (api.DbTester, error) {
	tester := &PostgresTester{}

	return tester, nil
}

func (tester *PostgresTester) GetQueryMarkers(length int) string {
	markerString := ""
	for markerIndex := 1; markerIndex <= length; markerIndex++ {
		markerString = markerString + fmt.Sprintf("$%d, ", markerIndex)
	}

	return markerString[:len(markerString)-2]
}

func (tester *PostgresTester) Connect(options api.DbTesterOptions) error {
	connectionOptions := map[string]string{}
	connectionOptions["dbname"] = options.DatabaseName

	if options.SslMode != "" {
		connectionOptions["sslmode"] = options.SslMode
	}

	if options.Port == "" {
		options.Port = "5432"
	}
	connectionOptions["port"] = options.Port

	connectionOptions["host"] = options.Host
	if options.Socket != "" {
		connectionOptions["host"] = options.Socket
	}

	if options.Username != "" && options.Password != "" {
		connectionOptions["user"] = options.Username
		connectionOptions["password"] = options.Password
	}

	connectionString := ""
	for key, value := range connectionOptions {
		connectionString = connectionString + fmt.Sprintf("%s=%s ",
			key,
			value)
	}

	if options.Debug {
		log.Printf("Connection string: %s", connectionString)
	}

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}

	if options.Debug {
		log.Printf("Connected to DB")
	}

	tester.Database = db
	tester.Debug = options.Debug

	return nil
}
