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
	for _, serverTLSSetting := range AllTLSSettings() {
		for _, socketType := range AllSocketTypes() {
			for _, sslMode := range AllSSLModes() {
				for _, publicCertStatus := range AllPublicCertStatuses() {
					for _, privateKeyStatus := range AllPrivateKeyStatuses() {
						for _, rootCertStatus := range AllRootCertStatuses() {

							connectionPort := ConnectionPort{
								// TODO: perhaps resolve this duplication of listener type
								SocketType: socketType,
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
									SocketType:       socketType,
									TLSSetting:       serverTLSSetting,
									SSLMode:          sslMode,
									RootCertStatus:   rootCertStatus,
									PrivateKeyStatus: privateKeyStatus,
									PublicCertStatus: publicCertStatus,
								},
								ConnectionPort: connectionPort,
							}


							handler.Credentials = append(
								handler.Credentials,
								// rootCertStatus
								rootCertStatus.toSecret(),
								//sslMode
								sslMode.toSecret(),
								//sslPrivateKeyTypeValue
								privateKeyStatus.toSecret(),
								//sslPublicCertTypeValue
								publicCertStatus.toSecret(),
								)
							// serverTLSSetting
							handler.Credentials = append(
								handler.Credentials,
								serverTLSSetting.toSecrets(CurrentDBConfig)...
							)

							// socketType
							switch socketType {
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
