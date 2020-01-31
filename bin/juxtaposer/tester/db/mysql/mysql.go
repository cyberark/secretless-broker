package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/cyberark/secretless-broker/bin/juxtaposer/tester/api"
	sql_db_tester "github.com/cyberark/secretless-broker/bin/juxtaposer/tester/db/sql"
)

type MySqlTester struct {
	sql_db_tester.SqlDatabaseTester
}

//[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]

func NewTester() (api.DbTester, error) {
	tester := &MySqlTester{}

	return tester, nil
}

func (tester *MySqlTester) GetQueryMarkers(length int) string {
	queryMarkers := strings.Split(strings.Repeat("?", length), "")
	return strings.Join(queryMarkers, ", ")
}

func (tester *MySqlTester) Connect(options api.DbTesterOptions) error {
	if options.Port == "" {
		options.Port = "3306"
	}

	host := fmt.Sprintf("tcp(%s:%s)", options.Host, options.Port)
	if strings.HasPrefix(options.Host, "/") {
		host = fmt.Sprintf("unix(%s)", options.Host)
	}

	authString := ""
	if options.Username != "" && options.Password != "" {
		authString = fmt.Sprintf("%s:%s@",
			url.QueryEscape(options.Username),
			url.QueryEscape(options.Password))
	}

	connectionString := fmt.Sprintf("%s%s/%s",
		authString,
		host,
		options.DatabaseName)

	if options.Debug {
		log.Printf("Connection string: %s", connectionString)
	}

	db, err := sql.Open("mysql", connectionString)
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
