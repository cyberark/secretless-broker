package test

import (
	"fmt"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)

// TODO: standardise on DB_PORT, DB_USER, DB_PASSWORD for flexibility
func sharedCredentials() []config.Variable {
	return []config.Variable{
		{
			Name:     "port",
			Provider: "env",
			ID:       "DB_PORT",
		},
		{
			Name:     "username",
			Provider: "env",
			ID:       "DB_USER",
		},
		{
			Name:     "password",
			Provider: "env",
			ID:       "DB_PASSWORD",
		},
	}
}


func GenerateConfigurations() (config.Config, LiveConfigurations) {
	// initialised with health-check listener and handler
	secretlessConfig := config.Config{
		Listeners: []config.Listener{
			{
				Debug:       true,
				Name:        "health-check",
				Protocol:    "mysql",
				Socket:      "/sock/mysql.sock",
			},
		},
		Handlers:  []config.Handler{
			{
				Name:         "health-check",
				ListenerName: "health-check",
				Debug:        true,
				Credentials:  []config.Variable{
					{
						Name:     "host",
						Provider: "literal",
						ID:       "health-check",
					},
					{
						Name:     "port",
						Provider: "literal",
						ID:       "3306",
					},
					{
						Name:     "username",
						Provider: "literal",
						ID:       "health-check",
					},
					{
						Name:     "password",
						Provider: "literal",
						ID:       "health-check",
					},
				},
			},
		},
	}

	liveConfigurations := make(LiveConfigurations, 0)

	// TODO: Create a utility xprod function similar to the one here:
	//     https://stackoverflow.com/questions/29002724/implement-ruby-style-cartesian-product-in-go
	// so we can avoid the nested for loops
	//
	// TODO: Remove "Value" suffixes -- no need for them, the lower case first letter
	// distinguishes them from the type itself, so it only degrades readability.
	portNumber := 3306
	for _, serverTLSTypeValue := range ServerTLSTypeValues() {
		for _, listenerTypeValue := range ListenerTypeValues() {
			for _, sslModeTypeValue := range SSlModeTypeValues() {
				for _, sslRootCertTypeValue := range SSLRootCertTypeValues() {
					listener := config.Listener{
						Name: fmt.Sprintf("listener_%v", portNumber),
						// TODO: grab value from envvar for flexibility
						Protocol: DBProtocol,
						Debug: true,
					}
					handler := config.Handler{
						Name: fmt.Sprintf("handler_%v", portNumber),
						Debug: true,
						ListenerName: fmt.Sprintf("listener_%v", portNumber),
						Credentials: sharedCredentials(),
					}
					liveConfiguration := LiveConfiguration{
						AbstractConfiguration: AbstractConfiguration{
							ListenerType:    listenerTypeValue,
							ServerTLSType:   serverTLSTypeValue,
							SSLModeType:     sslModeTypeValue,
							SSLRootCertType: sslRootCertTypeValue,
						},
					}

					// sslRootCertTypeValue
					handler.Credentials = append(handler.Credentials, sslRootCertTypeValue.toConfigVariable())

					//sslModeTypeValue
					// TODO: Make this same "toConfigVariable" refactoring for the other types
					handler.Credentials = append(handler.Credentials, sslModeTypeValue.toConfigVariable())

					// serverTLSTypeValue
					hostVariable := config.Variable{
						Name:     "host",
						Provider: "env",
						ID: 	  string(serverTLSTypeValue),
					}
					handler.Credentials = append(handler.Credentials, hostVariable)
					// serverTLSTypeValue
					// TODO: postgres and mysql should both have address
					// This is a hack and is forcing DB_HOST_NO_TLS
					// and DB_HOST_TLS to have host and port
					addressVariable := config.Variable{
						Name:     "address",
						Provider: "env",
						ID: 	  string(serverTLSTypeValue),
					}
					handler.Credentials = append(handler.Credentials, addressVariable)

					// listenerTypeValue
					switch listenerTypeValue {
					case TCP:
						listener.Address = fmt.Sprintf("0.0.0.0:%v", portNumber)
						liveConfiguration.port = fmt.Sprintf("%v", portNumber)
					case Socket:
						socket := fmt.Sprintf("/sock/db%v.sock", portNumber)
						listener.Socket = socket
						liveConfiguration.socket = socket
					}

					secretlessConfig.Listeners = append(secretlessConfig.Listeners, listener)
					secretlessConfig.Handlers = append(secretlessConfig.Handlers, handler)

					liveConfigurations = append(liveConfigurations, liveConfiguration)

					portNumber++
				}
			}
		}
	}

	return secretlessConfig, liveConfigurations
}
