module github.com/jenkins-x-apps/jx-app-sonar-scanner

go 1.12

require (
	github.com/antham/gommit v2.2.0+incompatible // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.3.3 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/jenkins-x/jx v1.3.1119
	github.com/kr/pty v1.1.5 // indirect
	github.com/magiconair/properties v1.8.0
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/petergtz/pegomock v2.7.0+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.5.1
	github.com/tektoncd/pipeline v0.3.2-0.20190508171555-91b53326d376 // indirect
	github.com/udhos/equalfile v0.3.0
	golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550 // indirect
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859 // indirect
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a // indirect
	golang.org/x/sync v0.0.0-20190423024810-112230192c58 // indirect
	golang.org/x/sys v0.0.0-20190616124812-15dcb6c0061f // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/appengine v1.5.0 // indirect
	gopkg.in/AlecAivazis/survey.v1 v1.8.3 // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.0.0-20190718062839-c8a0b81cb10e // indirect
	k8s.io/apimachinery v0.0.0-20190717022731-0bb8574e0887 // indirect
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)

replace github.com/heptio/sonobuoy => github.com/jenkins-x/sonobuoy v0.11.7-0.20190318120422-253758214767

replace k8s.io/api => k8s.io/api v0.0.0-20181128191700-6db15a15d2d3

replace k8s.io/metrics => k8s.io/metrics v0.0.0-20181128195641-3954d62a524d

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190122181752-bebe27e40fb7

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190528110200-4f3abb12cae2

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20181128195303-1f84094d7e8e

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v21.1.0+incompatible

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v10.15.5+incompatible

replace github.com/banzaicloud/bank-vaults => github.com/banzaicloud/bank-vaults v0.0.0-20190508130850-5673d28c46bd
