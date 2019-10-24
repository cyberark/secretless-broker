package main

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/stretchr/testify/assert"
)

func TestMSSQLConnector(t *testing.T) {
	t.Run("Can connect to MSSQL through Secretless", func(t *testing.T) {
		port := 2223
		server := "secretless"

		// Open the connection
		connString := fmt.Sprintf("server=%s;port=%d;encrypt=disable", server, port)
		conn, err := sql.Open("mssql", connString)
		assert.Nil(t, err, "you can open the db connection")
		if err != nil {
			return
		}
		defer conn.Close()

		// Prepare the test query
		testInt := 1
		testString := "abc"
		stmt, err := conn.Prepare(fmt.Sprintf("select %d, '%s'", testInt, testString))
		assert.Nil(t, err, "you can prepare the statement")
		if err != nil {
			return
		}
		defer stmt.Close()

		// Execute the query
		row := stmt.QueryRow()
		var actualInt int64
		var actualString string
		err = row.Scan(&actualInt, &actualString)
		assert.Nil(t, err, "you can read the queried values")
		if err != nil {
			return
		}

		// Test the returned values
		assert.EqualValues(t, testInt, actualInt)
		assert.EqualValues(t, testString, actualString)
	})

	t.Run("Cannot connect directly to MSSQL", func(t *testing.T) {
		port := 1433
		server := "mssql"

		// Open the connection
		connString := fmt.Sprintf("server=%s;port=%d;encrypt=disable", server, port)
		conn, err := sql.Open("mssql", connString)
		assert.Nil(t, err, "you can open the db connection")
		if err != nil {
			return
		}
		defer conn.Close()

		// Prepare the test query - should fail because the connection is not authenticated
		testInt := 1
		testString := "abc"
		_, err = conn.Prepare(fmt.Sprintf("select %d, '%s'", testInt, testString))
		assert.NotNil(t, err, "Prepare fails without a successful login")
		assert.Contains(t, err.Error(), "Login failed")
	})
}
