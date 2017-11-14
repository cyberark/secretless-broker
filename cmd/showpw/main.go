package main

import (
  "log"
  "os"

  "github.com/kgilpin/secretless/internal/pkg/keychain"
)

func main() {
  service := os.Args[1]
  account := os.Args[2]

  log.Print(keychain.GetGenericPassword(service, account))
}
