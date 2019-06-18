package sql

import (
	"database/sql"
	"fmt"
	"log"
)

type SqlDatabaseTester struct {
	Database *sql.DB
	Debug    bool
}

func (tester *SqlDatabaseTester) Query(query string, args ...interface{}) error {
	_, err := tester.QueryRows("", query, args...)
	return err
}

func (tester *SqlDatabaseTester) QueryRows(fieldName string,
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
		log.Printf("ERROR sql: Could not execute query!")
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

func (tester *SqlDatabaseTester) Shutdown() error {
	if tester.Debug {
		log.Println("Shutting down database connection...")
	}

	if tester.Database == nil {
		log.Println("WARN: Closing an already-closed connection!")
		return nil
	}

	err := tester.Database.Close()
	tester.Database = nil

	return err
}
