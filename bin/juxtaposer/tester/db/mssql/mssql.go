package mssql

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"

	"github.com/cyberark/secretless-broker/bin/juxtaposer/tester/api"
	sql_db_tester "github.com/cyberark/secretless-broker/bin/juxtaposer/tester/db/sql"
)

// MssqlTester is a wrapping struct around the basic SQL tester
type MssqlTester struct {
	sql_db_tester.SqlDatabaseTester
	databaseEnsured bool
}

// NewTester creates a new instance of the MSSQL DB tester
func NewTester() (api.DbTester, error) {
	tester := &MssqlTester{}

	return tester, nil
}

// sqlserver://username:password@host:port?database=master&param2=value

// GetQueryMarkers returns part of the query string that will be paramerized as it's
// different between databases. In this case, the params are defined using `@p<num>`.
func (tester *MssqlTester) GetQueryMarkers(length int) string {
	markers := []string{}
	for markerIndex := 1; markerIndex <= length; markerIndex++ {
		markers = append(markers, fmt.Sprintf("@p%d", markerIndex))
	}
	return strings.Join(markers, ", ")
}

// Connect is used to initialize a testing connection to the SQL database
func (tester *MssqlTester) Connect(options api.DbTesterOptions) error {
	if options.Port == "" {
		options.Port = "1433"
	}

	if options.Socket != "" {
		return fmt.Errorf("mssql driver doesn't support socket files")
	}

	query := url.Values{}
	query.Add("app name", "Juxtaposer")

	if !tester.databaseEnsured {
		err := tester.ensureDbExists(options, query)
		if err != nil {
			return err
		}
	}

	// The existence of the database has been ensured, so it's safe to
	// include it in the query header.
	query.Add("database", options.DatabaseName)
	db, err := tester.openDb(options, query)
	if err != nil {
		return err
	}

	tester.Database = db
	tester.Debug = options.Debug

	return nil
}

func (tester *MssqlTester) ensureDbExists(options api.DbTesterOptions, query url.Values) error {
	db, err := tester.openDb(options, query)
	if err != nil {
		return err
	}
	defer db.Close()

	log.Printf("Creating database (if it doesn't exist)...")
	createDbStmt := fmt.Sprintf(`
		IF NOT EXISTS (SELECT name FROM master.dbo.sysdatabases WHERE name = N'%s')
			CREATE DATABASE %s`,
		options.DatabaseName,
		options.DatabaseName)

	log.Printf("query string: %s\n", createDbStmt)
	_, err = db.Exec(createDbStmt)
	if err == nil {
		tester.databaseEnsured = true
	}
	return err
}

func (tester *MssqlTester) openDb(options api.DbTesterOptions, query url.Values) (*sql.DB, error) {
	connStringURL := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(options.Username, options.Password),
		Host:     fmt.Sprintf("%s:%s", options.Host, options.Port),
		RawQuery: query.Encode(),
	}

	connectionString := connStringURL.String()

	if options.Debug {
		log.Printf("Connection string: %s", connectionString)
	}

	db, err := sql.Open("sqlserver", connectionString)
	if (err == nil) && options.Debug {
		log.Printf("Connected to DB")
	}
	return db, err
}
