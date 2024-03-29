package pkg

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/cyberark/secretless-broker/test/util/testutil"
)

// RunQuery constructs a mysql cmdline command (which includes a sample query)
// for the options and credentials specified by the arguments.  It then executes
// the command, and returns its output.
func RunQuery(
	clientConfig testutil.ClientConfiguration,
	connectPort testutil.ConnectionPort,
) (string, error) {
	args := []string{"-e", "select count(*) from testdb.test"}

	if clientConfig.SSL {
		args = append(args, "--ssl-mode", "VERIFY_CA")
	}
	if clientConfig.Username != "" {
		args = append(args, fmt.Sprintf("--user=%s", clientConfig.Username))
	}
	if clientConfig.Password != "" {
		args = append(args, fmt.Sprintf("--password=%s", clientConfig.Password))
	}
	switch connectPort.SocketType {
	case testutil.TCP:
		args = append(args, fmt.Sprintf("--host=%s", connectPort.Host()))
		args = append(args, fmt.Sprintf("--port=%s", connectPort.ToPortString()))
	case testutil.Socket:
		args = append(args, fmt.Sprintf("--socket=%s", connectPort.ToSocketPath()))
	default:
		panic("Listener Type can only be TCP or Socket")
	}

	// Pre command logs
	println("")
	println("---<< EXECUTED")
	println(strings.Join(append([]string{"mysql"}, args...), " "))

	cmdOut, err := exec.Command("mysql", args...).CombinedOutput()

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
