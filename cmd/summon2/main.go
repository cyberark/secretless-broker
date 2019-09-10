package main

import (
	"fmt"
	"os"

	"os/exec"
	"syscall"

	"github.com/cyberark/secretless-broker/internal/summon/command"
)

func main() {
	if err := command.RunCLI(os.Args, os.Stdout); err != nil {
		code, err := returnStatusOfError(err)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(127)
		}

		os.Exit(code)
	}
}

// TODO: I am not sure what this is for
// It was brought over from the Summon code base.
func returnStatusOfError(err error) (int, error) {
	if eerr, ok := err.(*exec.ExitError); ok {
		if ws, ok := eerr.Sys().(syscall.WaitStatus); ok {
			if ws.Exited() {
				return ws.ExitStatus(), nil
			}
		}
	}
	return 0, err
}
