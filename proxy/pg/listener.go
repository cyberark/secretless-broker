package pg

import (
  "log"
  "net"

  "github.com/kgilpin/secretless/config"
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
      for _, handler := range self.Handlers {
        listener := handler.Listener
        if listener == "" {
          listener = handler.Name
        }

        if listener == self.Config.Name {
          handler := &Handler{Config: handler, Client: client}
          handler.Run()
        }
      }
    }
  }
}
