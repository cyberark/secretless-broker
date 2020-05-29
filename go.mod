module github.com/cyberark/secretless-broker

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/aws/aws-sdk-go v1.15.79
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/codegangsta/cli v1.20.0
	github.com/containerd/containerd v1.3.2 // indirect
	github.com/cyberark/conjur-api-go v0.5.2
	github.com/cyberark/conjur-authn-k8s-client v0.16.1
	github.com/cyberark/summon v0.7.0
	github.com/denisenkom/go-mssqldb v0.0.0-20191001013358-cfbb681360f0
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.4.2-0.20191231165639-e6f6c35b7902
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/fsnotify/fsnotify v1.4.7
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible
	github.com/gocql/gocql v0.0.0-20200505093417-effcbd8bcf0e
	github.com/google/btree v1.0.0 // indirect
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/gotestyourself/gotestyourself v1.4.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/vault/api v1.0.2
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/joho/godotenv v1.2.0
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/lib/pq v0.0.0-20180123210206-19c8e9ad0095
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/pkg/profile v1.2.1
	github.com/prometheus/client_golang v1.2.1 // indirect
	github.com/smartystreets/goconvey v0.0.0-20190731233626-505e41936337
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.3.0
	golang.org/x/crypto v0.0.0-20190510104115-cbcb75029529
	google.golang.org/appengine v1.4.0 // indirect
	gopkg.in/yaml.v2 v2.2.2
	gotest.tools v1.4.0 // indirect
	k8s.io/api v0.0.0-20180712090710-2d6f90ab1293
	k8s.io/apiextensions-apiserver v0.0.0-20180808065829-408db4a50408
	k8s.io/apimachinery v0.0.0-20180621070125-103fd098999d
	k8s.io/client-go v0.0.0-20180806134042-1f13a808da65
)

replace github.com/denisenkom/go-mssqldb => ./third_party/go-mssqldb

replace github.com/gocql/gocql => ./third_party/gocql

// 2/19/2019: cert on honnef.co -- one of grpc's dependencies -- expired.
// This is our fix:
replace honnef.co/go/tools => github.com/dominikh/go-tools v0.0.1-2019.2.3

go 1.13
