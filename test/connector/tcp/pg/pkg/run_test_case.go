package pkg

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/cyberark/secretless-broker/test/util/testutil"
)

const jdbcJARPath = "/secretless/test/util/jdbc/jdbc.jar"

// RunQuery runs a simply test query for the given client configuration and port.
func RunQuery(
	clientConfig testutil.ClientConfiguration,
	connectPort testutil.ConnectionPort,
) (string, error) {
	args := []string{"-c", "select count(*) from test.test"}
	connectionParams := []string{"dbname=postgres"}

	sslmode := "disable"
	if clientConfig.SSL {
		sslmode = "require"
	}
	connectionParams = append(connectionParams, fmt.Sprintf("sslmode=%s", sslmode))

	args = append(args, fmt.Sprintf("--username=%s", clientConfig.Username))
	connectionParams = append(connectionParams, fmt.Sprintf("password=%s", clientConfig.Password))

	args = append(args, fmt.Sprintf("--port=%s", connectPort.ToPortString()))

	var host string
	switch connectPort.SocketType {
	case testutil.TCP:
		host = connectPort.Host()
	case testutil.Socket:
		host = connectPort.ToSocketDir()
	default:
		panic("Listener Type can only be TCP or Socket")
	}
	args = append(args, fmt.Sprintf("--host=%s", host))

	// join connection params
	args = append(args, strings.Join(connectionParams, " "))

	// Pre command logs
	println("")
	println("---<< EXECUTED")
	println(strings.Join(append([]string{"psql"}, args...), " "))

	cmdOut, err := exec.Command("psql", args...).CombinedOutput()

	// Post command logs
	//TODO: deal with verbose
	if testutil.Verbose {
		if err != nil {
			println("--->> RESULTS")
			println("----- ERROR: ")
			println(err.Error())
		}
		println("----- OUTPUT: ")
		println(string(cmdOut))
	}
	println("---<< END")
	println("")

	return string(cmdOut), err
}

// RunQuery runs a simply test query for the given client configuration and port.
func RunJDBCQuery(
	clientConfig testutil.ClientConfiguration,
	connectPort testutil.ConnectionPort,
) (string, error) {

	args := []string{
		"-jar", jdbcJARPath,
		"-d", "postgres",
		"-m", "postgresql",
		"-h", fmt.Sprintf("%s:%d", connectPort.Host(), connectPort.Port),
		"-U", clientConfig.Username,
		"-P", clientConfig.Password,
		"select count(*) from test.test",
	}

	// Pre command logs
	println("")
	println("---->> ARGS")
	// deepcode ignore ClearTextLogging: This is a test file
	fmt.Println(args)

	println("---<< EXECUTED")
	cmdOut, err := exec.Command("java", args...).CombinedOutput()

	// Post command logs
	//TODO: deal with verbose
	if testutil.Verbose {
		if err != nil {
			println("--->> RESULTS")
			println("----- ERROR: ")
			println(err.Error())
		}
		println("----- OUTPUT: ")
		println(string(cmdOut))
	}
	println("---<< END")
	println("")

	return string(cmdOut), err
}
