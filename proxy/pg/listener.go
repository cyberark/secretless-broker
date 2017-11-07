package pg

import (
  "fmt"
  "log"
  "net"

  "github.com/kgilpin/secretless/config"
  "github.com/kgilpin/secretless/connect"
  "github.com/kgilpin/secretless/protocol"
)

type Listener struct {
  Config   config.Listener
  Handlers []config.Handler
  Listener net.Listener
}

func (self *Listener) Listen() {
  for {
    if client, err := self.Listener.Accept(); err != nil {
      log.Println(err)
      continue
    } else {
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

      if selectedHandler != nil {
        handler := &Handler{Config: *selectedHandler, Client: client}
        handler.Run()        
      } else {
        pgError := protocol.Error{
          Severity: protocol.ErrorSeverityFatal,
          Code:     protocol.ErrorCodeInternalError,
          Message:  fmt.Sprintf("No handler found for listener %s", self.Config.Name),
        }
        connect.Send(client, pgError.GetMessage())
      }
    }
  }
}
