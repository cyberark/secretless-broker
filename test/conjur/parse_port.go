package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	text := scanner.Text()
	tokens := strings.SplitN(text, ":", 2)
	if len(tokens) == 1 {
		fmt.Println(tokens[0])
	} else {
		fmt.Println(tokens[1])
	}
}
