package sshagent

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"strconv"

	"golang.org/x/crypto/ssh/agent"

	"github.com/conjurinc/secretless/internal/app/secretless/variable"
	"github.com/conjurinc/secretless/internal/pkg/provider"
	"github.com/conjurinc/secretless/pkg/secretless/config"
)

type Handler struct {
	Providers  []provider.Provider
	Config     config.Handler
	Connection net.Conn
}

func parseKey(pemStr string) (rawkey interface{}, err error) {
	pemBytes := []byte(pemStr)
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		err = fmt.Errorf("Failed to decode PEM block")
		return
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		rawkey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		rawkey, err = x509.ParseECPrivateKey(block.Bytes)
	default:
		err = fmt.Errorf("Unsupported key type %q", block.Type)
	}
	return
}

// LoadKeys loads the keys configured for this keyring handler.
func (h *Handler) LoadKeys(keyring agent.Agent) (err error) {
	var valuesPtr *map[string]string

	if valuesPtr, err = variable.Resolve(h.Providers, h.Config.Credentials); err != nil {
		return
	}

	values := *valuesPtr
	if h.Config.Debug {
		log.Printf("ssh-agent credential values : %s", values)
	}

	key := agent.AddedKey{}

	if rsa, ok := values["rsa"]; ok {
		key.PrivateKey, err = parseKey(rsa)
		if err != nil {
			return
		}
	}
	if ecdsa, ok := values["ecdsa"]; ok {
		key.PrivateKey, err = parseKey(ecdsa)
		if err != nil {
			return
		}
	}
	if comment, ok := values["comment"]; ok {
		key.Comment = comment
	}
	if lifetime, ok := values["lifetime"]; ok {
		var lt uint64
		lt, err = strconv.ParseUint(lifetime, 10, 32)
		if err != nil {
			return
		}
		key.LifetimeSecs = uint32(lt)
	}
	if confirm, ok := values["confirm"]; ok {
		key.ConfirmBeforeUse, err = strconv.ParseBool(confirm)
		if err != nil {
			return
		}
	}

	if h.Config.Debug {
		log.Printf("ssh-agent adding key : %s", key)
	}

	err = keyring.Add(key)
	return
}
