package pkg

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
			ID:       "MYSQL_PORT",
		},
		{
			Name:     "username",
			Provider: "literal",
			ID:       "testuser",
		},
		{
			Name:     "password",
			Provider: "env",
			ID:       "MYSQL_PASSWORD",
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

	portNumber := 3306
	for _, serverTLSTypeValue := range ServerTLSTypeValues() {
		for _, listenerTypeValue := range ListenerTypeValues() {
			for _, sslModeTypeValue := range SSlModeTypeValues() {
				for _, sslRootCertTypeValue := range SSLRootCertTypeValues() {
					listener := config.Listener{
						Name: fmt.Sprintf("listener_%v", portNumber),
						// TODO: grab value from envvar for flexibility
						Protocol: "mysql",
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
					if sslRootCertTypeValue != Undefined {
						sslRootCertVariable := config.Variable{
							Name:     "sslrootcert",
							Provider: "literal",
							ID:		   string(sslRootCertTypeValue),
						}
						handler.Credentials = append(handler.Credentials, sslRootCertVariable)
					}

					//sslModeTypeValue
					if sslModeTypeValue != Default {
						sslModeVariable := config.Variable{
							Name:     "sslmode",
							Provider: "literal",
							ID:		   string(sslModeTypeValue),
						}
						handler.Credentials = append(handler.Credentials, sslModeVariable)
					}

					// serverTLSTypeValue
					hostVariable := config.Variable{
						Name:     "host",
						Provider: "literal",
						ID: 	  string(serverTLSTypeValue),
					}
					handler.Credentials = append(handler.Credentials, hostVariable)

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
