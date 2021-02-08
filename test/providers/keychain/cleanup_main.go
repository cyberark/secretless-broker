package main

import "fmt"

// The main package simply cleans up the test fixtures
func main() {
	cleanup()
	fmt.Println("Cleanup: Removed test secrets from Keychain.")
}
