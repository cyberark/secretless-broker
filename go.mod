module github.com/cyberark/secretless-broker

require (
	github.com/Microsoft/go-winio v0.4.17 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/aws/aws-sdk-go v1.15.79
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/codegangsta/cli v1.20.0
	github.com/containerd/containerd v1.4.11 // indirect
	github.com/cyberark/conjur-api-go v0.5.2
	github.com/cyberark/conjur-authn-k8s-client v0.19.1
	github.com/cyberark/summon v0.7.0
	github.com/denisenkom/go-mssqldb v0.0.0-20191001013358-cfbb681360f0
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.4.2-0.20191231165639-e6f6c35b7902
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/vault/api v1.0.2
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/joho/godotenv v1.2.0
	github.com/keybase/go-keychain v0.0.0-20201121013009-976c83ec27a6
	github.com/lib/pq v0.0.0-20180123210206-19c8e9ad0095
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pkg/profile v1.2.1
	github.com/smartystreets/goconvey v1.6.4
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	google.golang.org/grpc v1.41.0 // indirect
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gotest.tools v2.2.0+incompatible // indirect
	k8s.io/api v0.22.3
	k8s.io/apiextensions-apiserver v0.22.3
	k8s.io/apimachinery v0.22.3
	k8s.io/client-go v0.22.3
)

replace github.com/denisenkom/go-mssqldb => ./third_party/go-mssqldb

// 2/19/2019: cert on honnef.co -- one of grpc's dependencies -- expired.
// This is our fix:
replace honnef.co/go/tools => github.com/dominikh/go-tools v0.0.1-2019.2.3

go 1.16
