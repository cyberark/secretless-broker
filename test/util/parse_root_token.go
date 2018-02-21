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
	if len(tokens) == 2 {
		fmt.Println(strings.TrimSpace(tokens[1]))
	} else {
		fmt.Errorf("Expected two colon delimited tokens in %s", text)
	}
}
