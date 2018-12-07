package pkg

import (
	"github.com/cyberark/secretless-broker/test/util/test"
	"os/exec"
	"strings"

	"github.com/smartystreets/goconvey/convey"
)


func RunQuery(flags []string) (string, error) {
	args := []string{"-e", "select count(*) from testdb.test"}

	for _, v := range flags {
		args = append(args, v)
	}

	// Pre command logs
	convey.Println("")
	convey.Println("---<< EXECUTED")
	convey.Println(strings.Join(append([]string{"mysql"}, args...), " "))

	cmdOut, err := exec.Command("mysql", args...).CombinedOutput()

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

