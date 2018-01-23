package sshagent

import (
  "log"
  "net"

  "golang.org/x/crypto/ssh/agent"

  "github.com/conjurinc/secretless/internal/pkg/provider"
  "github.com/conjurinc/secretless/pkg/secretless/config"
)

type Listener struct {
  Config    config.Listener
  Handlers  []config.Handler
  Providers []provider.Provider
  Listener  net.Listener
}

func (self *Listener) Listen() {
  // Serve the first Handler which is attached to this listener
  var selectedHandler *config.Handler
  for _, handler := range self.Handlers {
    listener := handler.Listener
    if listener == "" {
      listener = handler.Name
    }

    if listener == self.Config.Name {
      selectedHandler = &handler
      break
    }
  }

  if selectedHandler == nil {
    log.Fatalf("No ssh-agent handler is available")
  }

  keyring := agent.NewKeyring()

  handler := &Handler{Providers: self.Providers, Config: *selectedHandler}
  handler.LoadKeys(keyring)

  for {
    nConn, err := self.Listener.Accept()
    if err != nil {
      log.Printf("Failed to accept incoming connection: ", err)
      return
    }

    if err := agent.ServeAgent(keyring, nConn); err != nil {
      log.Printf("Error serving agent : %s", err)
    }
  }
}
