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
	"github.com/conjurinc/secretless/pkg/secretless/config"
)

// Handler implements an ssh-agent which holds a single key.
type Handler struct {
	Config     config.Handler
	Connection net.Conn
}

func parseKey(pemStr []byte) (rawkey interface{}, err error) {
	block, _ := pem.Decode(pemStr)
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

// LoadKeys loads the key configured for this ssh-agent.
//
// The Handler Credentials should provide a key as "rsa" or "ecdsa".
// "comment", "lifetime", and "confirm" are optional parameters.
func (h *Handler) LoadKeys(keyring agent.Agent) (err error) {
	var values map[string][]byte

	if values, err = variable.Resolve(h.Config.Credentials); err != nil {
		return
	}

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

	// TODO: comment, lifetime and confirm aren't credentials.
	// Maybe Handler needs a mechanism for these types of non-secret configuration options.
	if comment, ok := values["comment"]; ok {
		key.Comment = string(comment)
	}
	if lifetime, ok := values["lifetime"]; ok {
		var lt uint64
		lt, err = strconv.ParseUint(string(lifetime), 10, 32)
		if err != nil {
			return
		}
		key.LifetimeSecs = uint32(lt)
	}
	if confirm, ok := values["confirm"]; ok {
		key.ConfirmBeforeUse, err = strconv.ParseBool(string(confirm))
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

// GetConfig implements secretless.Handler
func (h *Handler) GetConfig() config.Handler {
	return h.Config
}

// GetClientConnection implements secretless.Handler
func (h *Handler) GetClientConnection() net.Conn {
	return h.Connection
}

// GetBackendConnection implements secretless.Handler
func (h *Handler) GetBackendConnection() net.Conn {
	return nil
}
