module github.com/cyberark/secretless-broker

require (
	github.com/asaskevich/govalidator v0.0.0-20180720115003-f9ffefc3facf // indirect
	github.com/aws/aws-sdk-go v1.15.79
	github.com/cenkalti/backoff v2.0.0+incompatible
	github.com/codegangsta/cli v1.20.0
	github.com/cyberark/conjur-api-go v0.5.2
	github.com/cyberark/conjur-authn-k8s-client v0.13.0
	github.com/cyberark/summon v0.7.0
	github.com/denisenkom/go-mssqldb v0.0.0-20191001013358-cfbb681360f0
	github.com/fsnotify/fsnotify v1.4.7
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible
	github.com/golang/groupcache v0.0.0-20180513044358-24b0969c4cb7 // indirect
	github.com/google/btree v0.0.0-20180813153112-4030bb1f1f0c // indirect
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/hashicorp/vault/api v1.0.2
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/hpcloud/tail v1.0.0 // indirect
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/joho/godotenv v1.2.0
	github.com/json-iterator/go v0.0.0-20180806060727-1624edc4454b // indirect
	github.com/lib/pq v0.0.0-20180123210206-19c8e9ad0095
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v0.0.0-20180718012357-94122c33edd3 // indirect
	github.com/onsi/ginkgo v1.6.0 // indirect
	github.com/onsi/gomega v1.4.1 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/profile v1.2.1
	github.com/prometheus/client_golang v0.9.2 // indirect
	github.com/sirupsen/logrus v1.0.6 // indirect
	github.com/smartystreets/goconvey v0.0.0-20190330032615-68dc04aab96a
	github.com/spf13/pflag v1.0.2 // indirect
	github.com/stretchr/testify v1.3.0
	golang.org/x/crypto v0.0.0-20190325154230-a5d413f7728c
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20180712090710-2d6f90ab1293
	k8s.io/apiextensions-apiserver v0.0.0-20180808065829-408db4a50408
	k8s.io/apimachinery v0.0.0-20180621070125-103fd098999d
	k8s.io/client-go v0.0.0-20180806134042-1f13a808da65
	k8s.io/kube-openapi v0.0.0-20180731170545-e3762e86a74c // indirect
)

replace github.com/denisenkom/go-mssqldb => github.com/cyberark/go-mssqldb v0.0.0-20191030142036-b5a965a47dd3

go 1.13
