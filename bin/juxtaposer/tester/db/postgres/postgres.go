package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"github.com/cyberark/secretless-broker/bin/juxtaposer/tester/api"
)

type PostgresTester struct {
	Database *sql.DB
	Debug    bool
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

func (tester *PostgresTester) Query(query string, args ...interface{}) error {
	_, err := tester.QueryRows("", query, args...)
	return err
}

func (tester *PostgresTester) QueryRows(fieldName string,
	query string, args ...interface{}) ([]string, error) {

	if tester.Debug {
		log.Printf("Query: %s", query)
		log.Print(args...)
	}

	if tester.Database == nil {
		return nil, fmt.Errorf("ERROR: Cannot query an unopened DB!")
	}

	rows, err := tester.Database.Query(query, args...)
	if err != nil {
		log.Printf("ERROR postgres: Could not execute query!")
		return nil, err
	}
	defer rows.Close()

	fieldValues := make([]string, 0)
	for rows.Next() {
		var fieldValue string
		if err := rows.Scan(&fieldValue); err != nil {
			return nil, err
		}
		fieldValues = append(fieldValues, fieldValue)
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return fieldValues, nil
}

func (tester *PostgresTester) Shutdown() error {
	if tester.Debug {
		log.Println("Shutting down database connection...")
	}

	tester.Database.Close()
	return nil
}
