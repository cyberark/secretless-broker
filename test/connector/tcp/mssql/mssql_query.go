package mssqltest

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"os/exec"
	"text/tabwriter"

	_ "github.com/denisenkom/go-mssqldb"

	"github.com/cyberark/secretless-broker/test/util/testutil"
)

type dbConfig struct {
	Host string
	Port int
	Username string
	Password string
	Database string
}

type dbQueryExecutor func(cfg dbConfig, query string) (string, error)

func defaultSecretlessDbConfig() dbConfig {
	return dbConfig{
		Host:     testutil.SecretlessHost,
		Port:     2223,
		Username: "dummy",
		Password: "dummy",
	}
}

// runs queries using sqlcmd
func sqlcmdExec(
	cfg dbConfig,
	query string,
) (string, error) {
	args := []string{
		"-S", fmt.Sprintf("%s,%d", cfg.Host, cfg.Port),
		"-U", cfg.Username,
		"-P", cfg.Password,
		"-Q", query,
	}

	if db := cfg.Database; db != "" {
		args = append(args, "-d", db)
	}

	out, err := exec.Command(
		"sqlcmd",
		args...
	).Output()

	if err != nil {
		if exitErrr, ok := err.(*exec.ExitError); ok {
			return "", errors.New(string(exitErrr.Stderr))
		}

		return "", err
	}

	return string(out), nil
}

// runs queries using go-mssqldb
func gomssqlExec(
	cfg dbConfig,
	query string,
) (string, error) {
	dsnString := fmt.Sprintf(
		"user id=%s;password=%s;server=%s;port=%d;encrypt=%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		"disable",
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
