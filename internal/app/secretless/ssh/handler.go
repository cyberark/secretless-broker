package ssh

import (
  "io"
  "log"

  raw_ssh "golang.org/x/crypto/ssh"

  "github.com/gliderlabs/ssh"

  "github.com/kgilpin/secretless/internal/app/secretless/variable"
  "github.com/kgilpin/secretless/internal/pkg/provider"
  "github.com/kgilpin/secretless/pkg/secretless/config"
)

type ServerConfig struct {
  Network       string
  Address       string
  ClientConfig  raw_ssh.ClientConfig
}

type Handler struct {
  Providers []provider.Provider
  Config    config.Handler
  Session   ssh.Session
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
    var hostKey raw_ssh.PublicKey
    if hostKey, err = raw_ssh.ParsePublicKey([]byte(hostKeyStr)); err != nil {
      log.Printf("Unable to parse public key: %v", err)
      return
    }
    config.ClientConfig.HostKeyCallback = raw_ssh.FixedHostKey(hostKey)
  } else {
    log.Printf("No hostKey specified; I will accept any host key!")
    config.ClientConfig.HostKeyCallback = raw_ssh.InsecureIgnoreHostKey()
  }

  if privateKeyStr, ok := values["privateKey"]; ok {
    var signer raw_ssh.Signer
    if signer, err = raw_ssh.ParsePrivateKey([]byte(privateKeyStr)); err != nil {
      log.Printf("Unable to parse private key: %v", err)
      return
    }
    config.ClientConfig.Auth = []raw_ssh.AuthMethod{
      raw_ssh.PublicKeys(signer),
    }
  }

  return
}

func (self *Handler) Run() {
  var err error
  var serverConfig ServerConfig

  if serverConfig, err = self.serverConfig(); err != nil {
    return
  }

  if client, err := raw_ssh.Dial(serverConfig.Network, serverConfig.Address, &serverConfig.ClientConfig); err != nil {
    log.Printf("Failed to dial SSH backend '%s': %s", serverConfig.Address, err)
    return
  }

  io.WriteString(self.Session, "Connected!\n")

  // Each ClientConn can support multiple interactive sessions,
  // represented by a Session.
  session, err := client.NewSession()
  if err != nil {
      log.Fatal("Failed to create session: ", err)
  }
  defer session.Close()

  go io.Copy(self.Session.Stdin, session.Stdin)
  go io.Copy(session.Stdout, self.Session.Stdout)
  go io.Copy(session.Stderr, self.Session.Stderr)
}
