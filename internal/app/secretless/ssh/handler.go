package ssh

import (
  "io"
  "log"

  "golang.org/x/crypto/ssh"

  "github.com/kgilpin/secretless/internal/app/secretless/variable"
  "github.com/kgilpin/secretless/internal/pkg/provider"
  "github.com/kgilpin/secretless/pkg/secretless/config"
)

type ServerConfig struct {
  Network       string
  Address       string
  ClientConfig  ssh.ClientConfig
}

type Handler struct {
  Providers []provider.Provider
  Config    config.Handler
  Channels  <-chan ssh.NewChannel
}

func (self *Handler) serverConfig() (config ServerConfig, err error) {
  var valuesPtr *map[string]string

  log.Printf("%s", self.Config.Credentials)

  if valuesPtr, err = variable.Resolve(self.Providers, self.Config.Credentials); err != nil {
    return
  }

  values := *valuesPtr
  if self.Config.Debug {
    log.Printf("SSH backend connection parameters: %s", values)
  }

  config.Network = "tcp"
  if address, ok := values["address"]; ok {
    config.Address = address
  }

  if user, ok := values["user"]; ok {
    config.ClientConfig.User = user
  }

  if hostKeyStr, ok := values["hostKey"]; ok {
    var hostKey ssh.PublicKey
    if hostKey, err = ssh.ParsePublicKey([]byte(hostKeyStr)); err != nil {
      log.Printf("Unable to parse public key: %v", err)
      return
    }
    config.ClientConfig.HostKeyCallback = ssh.FixedHostKey(hostKey)
  } else {
    log.Printf("No hostKey specified; I will accept any host key!")
    config.ClientConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
  }

  if privateKeyStr, ok := values["privateKey"]; ok {
    var signer ssh.Signer
    if signer, err = ssh.ParsePrivateKey([]byte(privateKeyStr)); err != nil {
      log.Printf("Unable to parse private key: %v", err)
      return
    }
    config.ClientConfig.Auth = []ssh.AuthMethod{
      ssh.PublicKeys(signer),
    }
  }

  return
}

func (self *Handler) Run() {
  var err error
  var serverConfig ServerConfig
  var server ssh.Conn

  if serverConfig, err = self.serverConfig(); err != nil {
    return
  }

  if server, err = ssh.Dial(serverConfig.Network, serverConfig.Address, &serverConfig.ClientConfig); err != nil {
    log.Printf("Failed to dial SSH backend '%s': %s", serverConfig.Address, err)
    return
  }

  // Service the incoming Channel channel.
  for newChannel := range self.Channels {
    serverChannel, serverRequests, err := server.OpenChannel(newChannel.ChannelType(), newChannel.ExtraData())
    if err != nil {
      sshError := err.(*ssh.OpenChannelError)
      if err := newChannel.Reject(sshError.Reason, sshError.Message); err != nil {
        log.Printf("Failed to send new channel rejection : %s", err)
      }
      return
    }

    clientChannel, clientRequests, err := newChannel.Accept()
    if err != nil {
      log.Printf("Failed to accept client channel : %s", err)
      serverChannel.Close()
      return
    }

    go func() {
      for clientRequest := range clientRequests {
        log.Printf("Client request : %s", clientRequest.Type)
        ok, err := serverChannel.SendRequest(clientRequest.Type, clientRequest.WantReply, clientRequest.Payload)
        if err != nil {
          log.Printf("Failed to send client request to server channel : %s", err)
        }
        if clientRequest.WantReply {
          log.Printf("Server reply is %v", ok)
        }
      }
    }()

    go func() {
      for serverRequest := range serverRequests {
        log.Printf("Server request : %s", serverRequest.Type)
        ok, err := clientChannel.SendRequest(serverRequest.Type, serverRequest.WantReply, serverRequest.Payload)
        if err != nil {
          log.Printf("Failed to send server request to client channel : %s", err)
        }
        if serverRequest.WantReply {
          log.Printf("Client reply is %v", ok)
        }
      }    
    }()
  
    go func() {
      for {
        data := make([]byte, 1024)
        len, err := clientChannel.Read(data)
        if err == io.EOF {
          log.Printf("Client channel is closed")
          serverChannel.Close()
          return
        }
        _, err = serverChannel.Write(data[0:len])
        if err != nil {
          log.Printf("Error writing %s bytes to server channel : %s", len, err)
        }
      }
      }()

    go func() {
      for {
        data := make([]byte, 1024)
        len, err := serverChannel.Read(data)
        if err == io.EOF {
          log.Printf("Server channel is closed")
          clientChannel.Close()
          return
        }
        _, err = clientChannel.Write(data[0:len])
        if err != nil {
          log.Printf("Error writing %s bytes to client channel : %s", len, err)
        }
      }
      }()


    // Move the data from the client to the server

    // Move the server responses to the client

    // clientChannel.Close()
    // serverChannel.Close()
  }
}
