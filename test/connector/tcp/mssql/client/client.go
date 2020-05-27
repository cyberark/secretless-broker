package client

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os/exec"
	"text/tabwriter"
	"time"

	_ "github.com/denisenkom/go-mssqldb" // register mssql driver
)

// jdbcJARPath is the path to jar file containing the jdbc client.
const jdbcJARPath = "/secretless/test/util/jdbc/jdbc.jar"

// SqlcmdExec runs a query by invoking sqlcmd
func SqlcmdExec(
	cfg Config,
	query string,
) (string, error) {
	args := []string{
		"-S", fmt.Sprintf("%s,%s", cfg.Host, cfg.Port),
		"-U", cfg.Username,
		"-P", cfg.Password,
		"-Q", query,
	}

	if cfg.ReadOnly == true {
		args = append(args, "-K", "ReadOnly")
	}

	if db := cfg.Database; db != "" {
		args = append(args, "-d", db)
	}

	out, err := exec.Command(
		"sqlcmd",
		args...,
	).Output()

	if err != nil {
		if exitErrr, ok := err.(*exec.ExitError); ok {
			return "", errors.New(string(exitErrr.Stderr))
		}

		return "", err
	}

	return string(out), nil
}

// PythonODBCExec runs a query by invoking python-odbc
func PythonODBCExec(
	cfg Config,
	query string,
) (string, error) {
	applicationintent := "readwrite"
	if cfg.ReadOnly {
		applicationintent = "readonly"
	}

	args := []string{
		"--server", fmt.Sprintf("%s,%s", cfg.Host, cfg.Port),
		"--username", cfg.Username,
		"--password", cfg.Password,
		"--query", query,
		"--application-intent", applicationintent,
	}

	if db := cfg.Database; db != "" {
		args = append(args, "--database", db)
	}

	out, err := exec.Command(
		"./client/odbc_client.py",
		args...,
	).Output()

	if err != nil {
		if exitErrr, ok := err.(*exec.ExitError); ok {
			return "", errors.New(string(exitErrr.Stderr))
		}

		return "", err
	}

	return string(out), nil
}

// JavaJDBCExec runs a query by invoking Java JDBC
// Jar modified from this source: http://jdbcsql.sourceforge.net/
func JavaJDBCExec(
	cfg Config,
	query string,
) (string, error) {
	args := []string{
		"-jar", jdbcJARPath,
		"-m", "mssql",
		"-h", fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		"-U", cfg.Username,
		"-P", cfg.Password,
	}

	// For JDBC, database is not optional. If empty, add the default MsSQL database
	if db := cfg.Database; db == "" {
		args = append(args, "-d", "tempdb")
	} else {
		args = append(args, "-d", db)
	}

	args = append(args, query)

	out, err := exec.Command(
		"java",
		args...,
	).Output()

	if err != nil {
		if exitErrr, ok := err.(*exec.ExitError); ok {
			return "", errors.New(string(exitErrr.Stderr))
		}

		return "", err
	}

	return string(out), nil
}

// GomssqlExec runs a query by invoking go-mssqldb
func GomssqlExec(
	cfg Config,
	query string,
) (string, error) {
	applicationIntent := "ReadWrite"
	if cfg.ReadOnly {
		applicationIntent = "ReadOnly"
	}

	dsnString := fmt.Sprintf(
		"user id=%s;password=%s;server=%s;port=%s;encrypt=%s;applicationintent=%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		"disable",
		applicationIntent,
	)

	if db := cfg.Database; db != "" {
		dsnString += fmt.Sprintf(";database=%s", db)
	}

	// Open the connection
	conn, err := sql.Open(
		"mssql",
		dsnString,
	)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	ctx, _ := context.WithDeadline(
		context.Background(),
		time.Now().Add(1*time.Second),
	)

	if query == "" {
		_, err := conn.Conn(ctx)
		return "", err
	}

	// Execute the query
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// Execute the query
	cols, err := rows.Columns()
	if err != nil {
		return "", err
	}

	rawResult := make([][]byte, len(cols))

	dest := make([]interface{}, len(cols)) // A temporary interface{} slice
	for i := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	w := new(tabwriter.Writer)
	buf := &bytes.Buffer{}
	w.Init(buf, 0, 0, 0, ' ', tabwriter.Debug|tabwriter.AlignRight)

	for rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			return "", nil
		}

		rowString := ""
		for _, raw := range rawResult {
			if raw == nil {
				rowString += "\\N"
			} else {
				rowString += string(raw)
			}

			rowString += "\t"
		}

		fmt.Fprintln(w, rowString)
	}

	w.Flush()

	return string(buf.Bytes()), err
}
