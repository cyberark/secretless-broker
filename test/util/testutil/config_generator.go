package testutil

import (
	config_v2 "github.com/cyberark/secretless-broker/pkg/secretless/config/v2"
)

// GenerateConfigurations returns a Secretless Config along with a comprehensive
// list of LiveConfigurations for use in tests.
// TODO: consider parametrising ConnectPort generator
func GenerateConfigurations() (config_v2.Config, LiveConfigurations) {
	// initialised with health-check listener and handler
	secretlessConfig := config_v2.Config{
		Services: []*config_v2.Service{
			{
				Debug:           true,
				Connector:       "mysql",
				ConnectorConfig: nil,
				Credentials: []*config_v2.Credential{
					{
						Name: "host",
						From: "literal",
						Get:  sampleDbConfig.HostWithTLS,
					},
					{
						Name: "username",
						From: "literal",
						Get:  sampleDbConfig.User,
					},
					{
						Name: "password",
						From: "literal",
						Get:  sampleDbConfig.Password,
					},
				},
				ListenOn: "unix:///sock/mysql.sock",
				Name:     "health-check",
			},
			{
				Debug:           true,
				Connector:       "pg",
				ConnectorConfig: nil,
				Credentials: []*config_v2.Credential{
					{
						Name: "host",
						From: "literal",
						Get:  "health-check",
					},
					{
						Name: "port",
						From: "literal",
						Get:  "3306",
					},
					{
						Name: "username",
						From: "literal",
						Get:  "health-check",
					},
					{
						Name: "password",
						From: "literal",
						Get:  "health-check",
					},
				},
				ListenOn: "tcp://0.0.0.0:5432",
				Name:     "pg-bench",
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
				for _, sslHost := range AllSSLHosts() {
					for _, publicCertStatus := range AllPublicCertStatuses() {
						for _, privateKeyStatus := range AllPrivateKeyStatuses() {
							for _, rootCertStatus := range AllRootCertStatuses() {
								for _, areAuthCredentialsInvalid := range AllAuthCredentialsInvalidity() {

									connectionPort := ConnectionPort{
										// TODO: perhaps resolve this duplication of listener type
										SocketType: socketType,
										Port:       portNumber,
									}

									name := "test_service_" + connectionPort.ToPortString()
									credentials := areAuthCredentialsInvalid.toSecrets()

									liveConfiguration := LiveConfiguration{
										AbstractConfiguration: AbstractConfiguration{
											SocketType:               socketType,
											TLSSetting:               serverTLSSetting,
											SSLHost:                  sslHost,
											SSLMode:                  sslMode,
											RootCertStatus:           rootCertStatus,
											PrivateKeyStatus:         privateKeyStatus,
											PublicCertStatus:         publicCertStatus,
											AuthCredentialInvalidity: areAuthCredentialsInvalid,
										},
										ConnectionPort: connectionPort,
									}

									credentials = append(
										credentials,
										// rootCertStatus
										rootCertStatus.toSecret(),
										//sslMode
										sslMode.toSecret(),
										//sslHost
										sslHost.toSecret(),
										//sslPrivateKeyTypeValue
										privateKeyStatus.toSecret(),
										//sslPublicCertTypeValue
										publicCertStatus.toSecret(),
									)
									// serverTLSSetting
									credentials = append(
										credentials,
										serverTLSSetting.toSecrets(sampleDbConfig)...,
									)

									// socketType
									address := ""
									switch socketType {
									case TCP:
										address = "tcp://0.0.0.0:" + connectionPort.ToPortString()
									case Socket:
										address = "unix://" + connectionPort.ToSocketPath()
									}

									svc := &config_v2.Service{
										Debug: true,
										// TODO: grab value from envvar for flexibility
										Connector:       sampleDbConfig.Protocol,
										ConnectorConfig: nil,
										Credentials:     credentials,
										ListenOn:        config_v2.NetworkAddress(address),
										Name:            name,
									}

									secretlessConfig.Services = append(
										secretlessConfig.Services,
										svc)
									liveConfigurations = append(liveConfigurations, liveConfiguration)

									portNumber++
								}
							}
						}
					}
				}
			}
		}
	}

	return secretlessConfig, liveConfigurations
}
