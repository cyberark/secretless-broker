package main

import (
	"fmt"
	"os"

	"github.com/conjurinc/secretless/internal/app/summon/command"
)

func main() {
	if err := command.RunCLI(os.Args, os.Stdout); err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}
