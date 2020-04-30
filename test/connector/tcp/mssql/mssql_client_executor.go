package mssqltest

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os/exec"
	"text/tabwriter"

	_ "github.com/denisenkom/go-mssqldb" // register mssql driver
)

// dbClientConfig is abstract and represents the configurations that apply to all
// clients, each dbClientExecutor translates the configuration into a form that makes
// sense for its client.
// e.g. Username, Database translate to the following command for sqlcmd:
//
// sqlcmd -d Database -U Username
//
type dbClientConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
	// This is in relation to what is generally referred to as Application Intent.
	// It can only take 2 values, ReadWrite or ReadOnly.
	ReadOnly bool
}

// dbClientExecutor represents the invocation of an MSSQL client. It takes two arguments,
// database client configuration (dbClientConfig) and query (string). It returns a string
// and an error; the string captures the success output and the error captures the failure.
//
// As an example, sqlcmdExec is of type dbClientExecutor. sqlcmdExec invokes the sqlcmd
// program using the arguments provided. An example invocation might look as follows:
//
// sqlcmd -d Database -U Username -Q query
//
type dbClientExecutor func(cfg dbClientConfig, query string) (string, error)

// clientResponse represents the response from calling a dbClientExecutor. It is composed
// of a string and an error; the string captures the success output and the error
// captures the failure.
type clientResponse struct {
	out string
	err error
}

// concurrentClientExec calls the dbClientExecutor concurrently, and returns a channel
// that can be waited on to get the client response.
func concurrentClientExec(
	executor dbClientExecutor,
	clientCfg dbClientConfig,
	query string,
) chan clientResponse {
	clientResChan := make(chan clientResponse)

	go func() {
		out, err := executor(
			clientCfg,
			query,
		)

		clientResChan <- clientResponse{
			out: out,
			err: err,
		}
	}()

	return clientResChan
}

const jdbcJARPath = "/secretless/test/util/jdbc/jdbc.jar"

// sqlcmdExec runs a query by invoking sqlcmd
func sqlcmdExec(
	cfg dbClientConfig,
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

// pythonODBCExec runs a query by invoking python-odbc
func pythonODBCExec(
	cfg dbClientConfig,
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
		"./odbc_client.py",
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

// javaJDBCExec runs a query by invoking Java JDBC
// Jar modified from this source: http://jdbcsql.sourceforge.net/
func javaJDBCExec(
	cfg dbClientConfig,
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

// gomssqlExec runs a query by invoking go-mssqldb
func gomssqlExec(
	cfg dbClientConfig,
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

	if query == "" {
		_, err := conn.Conn(context.Background())
		return "", err
	}

	// Execute the query
	rows, err := conn.Query(query)
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
