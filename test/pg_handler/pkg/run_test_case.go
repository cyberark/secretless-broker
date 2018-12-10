package pkg

import (
	"fmt"
	"github.com/cyberark/secretless-broker/test/util/test"
	"os/exec"
	"strings"

	"github.com/smartystreets/goconvey/convey"
)


func RunQuery(clientConfig test.ClientConfiguration, connectPort test.ConnectionPort) (string, error) {
	args := []string{"-c", "select count(*) from test.test"}
	connectionParams := []string{"dbname=postgres"}

	sslmode := "disable"
	if clientConfig.SSL != nil && *clientConfig.SSL {
		sslmode = "require"
	}
	connectionParams = append(connectionParams, fmt.Sprintf("sslmode=%s", sslmode))

	if clientConfig.Username != nil {
		args = append(args, fmt.Sprintf("--username=%s", *clientConfig.Username))
	}
	if clientConfig.Password != nil {
		connectionParams = append(connectionParams, fmt.Sprintf("password=%s", *clientConfig.Password))
	}

	args = append(args, fmt.Sprintf("--port=%s", connectPort.ToPortString()))

	var host string
	switch connectPort.ListenerType {
	case test.TCP:
		host = connectPort.Host()
	case test.Socket:
		host = connectPort.ToSocketDir()
	default:
		panic("Listener Type can only be TCP or Socket")
	}
	args = append(args, fmt.Sprintf("--host=%s", host))

	// join connection params
	args = append(args, strings.Join(connectionParams, " "))


	// Pre command logs
	convey.Println("")
	convey.Println("---<< EXECUTED")
	convey.Println(strings.Join(append([]string{"psql"}, args...), " "))

	cmdOut, err := exec.Command("psql", args...).CombinedOutput()

	// Post command logs
	//TODO: deal with verbose
	if test.Verbose {
		if err != nil {
			convey.Println("--->> RESULTS")
			convey.Println("----- ERROR: ")
			convey.Println(err.Error())
		}
		convey.Println("----- OUTPUT: ")
		convey.Println(string(cmdOut))
	}
	convey.Println("---<< END")
	convey.Println("")

	return string(cmdOut), err
}

type ConnectionParams struct {
	Username string
	Password string
	SSLMode string
}
