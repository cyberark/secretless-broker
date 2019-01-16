package test

import (
	"fmt"
	"github.com/cyberark/secretless-broker/pkg/secretless/config"
)

// TODO: standardise on DB_PORT, DB_USER, DB_PASSWORD for flexibility
func sharedCredentials() []config.StoredSecret {
	return []config.StoredSecret{
		{
			Name:     "username",
			Provider: "literal",
			ID:       CurrentDBConfig.User,
		},
		{
			Name:     "password",
			Provider: "literal",
			ID:       CurrentDBConfig.Password,
		},
	}
}


// TODO: consider parametrising ConnectPort generator
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
			{
				Debug:       true,
				Name:        "pg-bench",
				Protocol:    "pg",
				Address:     "0.0.0.0:5432",
			},
		},
		Handlers:  []config.Handler{
			{
				Name:         "health-check",
				ListenerName: "health-check",
				Debug:        true,
				Credentials:  []config.StoredSecret{
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
			{
				Name:         "pg-bench",
				ListenerName: "pg-bench",
				Debug:        true,
				Credentials:  []config.StoredSecret{
					{
						Name:     "address",
						Provider: "literal",
						ID:       fmt.Sprintf("%s:5432", CurrentDBConfig.HostWithTLS),
					},
					{
						Name:     "username",
						Provider: "literal",
						ID:       CurrentDBConfig.User,
					},
					{
						Name:     "password",
						Provider: "literal",
						ID:       CurrentDBConfig.Password,
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
	portNumber := 3307
	for _, serverTLSTypeValue := range AllServerTLSSettings() {
		for _, listenerTypeValue := range AllSocketTypes() {
			for _, sslModeTypeValue := range SSlModeTypeValues() {
				for _, sslPublicCertValue := range SSLPublicCertTypeValues() {
					for _, sslPrivateKeyValue := range SSLPrivateKeyTypeValues() {
						for _, sslRootCertTypeValue := range SSLRootCertTypeValues() {
							connectionPort := ConnectionPort{
								// TODO: perhaps resolve this duplication of listener type
								SocketType: listenerTypeValue,
								Port:       portNumber,
							}

							listener := config.Listener{
								Name: "listener_" + connectionPort.ToPortString(),
								// TODO: grab value from envvar for flexibility
								Protocol: CurrentDBConfig.Protocol,
								Debug:    true,
							}
							handler := config.Handler{
								Name:         "handler_" + connectionPort.ToPortString(),
								Debug:        true,
								ListenerName: "listener_" + connectionPort.ToPortString(),
								Credentials:  sharedCredentials(),
							}
							liveConfiguration := LiveConfiguration{
								AbstractConfiguration: AbstractConfiguration{
									SocketType:        listenerTypeValue,
									ServerTLSSetting:  serverTLSTypeValue,
									SSLModeType:       sslModeTypeValue,
									SSLRootCertType:   sslRootCertTypeValue,
									SSLPrivateKeyType: sslPrivateKeyValue,
									SSLPublicCertType: sslPublicCertValue,
								},
								ConnectionPort: connectionPort,
							}


							handler.Credentials = append(
								handler.Credentials,
								// sslRootCertTypeValue
								sslRootCertTypeValue.toConfigVariable(),
								//sslModeTypeValue
								sslModeTypeValue.toConfigVariable(),
								//sslPrivateKeyTypeValue
								sslPrivateKeyValue.toConfigVariable(),
								//sslPublicCertTypeValue
								sslPublicCertValue.toConfigVariable(),
								)
							// serverTLSTypeValue
							handler.Credentials = append(
								handler.Credentials,
								serverTLSTypeValue.toSecrets(CurrentDBConfig)...
							)

							// listenerTypeValue
							switch listenerTypeValue {
							case TCP:
								listener.Address = "0.0.0.0:" + connectionPort.ToPortString()
							case Socket:
								listener.Socket = connectionPort.ToSocketPath()
							}

							secretlessConfig.Listeners = append(secretlessConfig.Listeners, listener)
							secretlessConfig.Handlers = append(secretlessConfig.Handlers, handler)

							liveConfigurations = append(liveConfigurations, liveConfiguration)

							portNumber++
						}
					}
				}
			}
		}
	}

	return secretlessConfig, liveConfigurations
}
