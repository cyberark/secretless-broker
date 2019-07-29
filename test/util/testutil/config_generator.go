package testutil

import (
	config_v1 "github.com/cyberark/secretless-broker/pkg/secretless/config/v1"
)

// GenerateConfigurations returns a Secretless Config along with a comprehensive
// list of LiveConfigurations for use in tests.
// TODO: consider parametrising ConnectPort generator
func GenerateConfigurations() (config_v1.Config, LiveConfigurations) {
	// initialised with health-check listener and handler
	secretlessConfig := config_v1.Config{
		Listeners: []config_v1.Listener{
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
		Handlers:  []config_v1.Handler{
			{
				Name:         "health-check",
				ListenerName: "health-check",
				Debug:        true,
				Credentials:  []config_v1.StoredSecret{
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
				Credentials:  []config_v1.StoredSecret{
					{
						Name:     "host",
						Provider: "literal",
						ID:       sampleDbConfig.HostWithTLS,
					},
					{
						Name:     "username",
						Provider: "literal",
						ID:       sampleDbConfig.User,
					},
					{
						Name:     "password",
						Provider: "literal",
						ID:       sampleDbConfig.Password,
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
							for _, areAuthCredentialsInvalid := range AllAuthCredentialsInvalidity() {

								connectionPort := ConnectionPort{
									// TODO: perhaps resolve this duplication of listener type
									SocketType: socketType,
									Port:       portNumber,
								}

								listener := config_v1.Listener{
									Name: "listener_" + connectionPort.ToPortString(),
									// TODO: grab value from envvar for flexibility
									Protocol: sampleDbConfig.Protocol,
									Debug:    true,
								}
								handler := config_v1.Handler{
									Name:         "handler_" + connectionPort.ToPortString(),
									Debug:        true,
									ListenerName: "listener_" + connectionPort.ToPortString(),
									// auth credentials
									Credentials: areAuthCredentialsInvalid.toSecrets(),
								}
								liveConfiguration := LiveConfiguration{
									AbstractConfiguration: AbstractConfiguration{
										SocketType:       socketType,
										TLSSetting:       serverTLSSetting,
										SSLMode:          sslMode,
										RootCertStatus:   rootCertStatus,
										PrivateKeyStatus: privateKeyStatus,
										PublicCertStatus: publicCertStatus,
										AuthCredentialInvalidity: areAuthCredentialsInvalid,
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
									serverTLSSetting.toSecrets(sampleDbConfig)...
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
	}

	return secretlessConfig, liveConfigurations
}
