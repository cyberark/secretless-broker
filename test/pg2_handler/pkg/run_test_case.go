package pkg

import (
	"fmt"
	"github.com/cyberark/secretless-broker/test/util/test"
	"os/exec"
	"strings"

	"github.com/smartystreets/goconvey/convey"
)


func RunQuery(flags []string) (string, error) {
	args := []string{"-c", "select count(*) from test.test"}

	for _, v := range flags {
		args = append(args, v)
	}

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

func (cp ConnectionParams) ToFlags() []string  {
	var flags []string

	var (
		SSLMode = "disable"
	)

	if cp.SSLMode != "" {
		SSLMode = cp.SSLMode
	}
	flags = append(flags,
		fmt.Sprintf("dbname=postgres sslmode=%s password=%s", SSLMode, cp.Password),
	)

	flags = append(flags, fmt.Sprintf("--username=%s", cp.Username))

	return flags
}
