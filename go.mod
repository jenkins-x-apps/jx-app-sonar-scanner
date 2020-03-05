module github.com/jenkins-x-apps/jx-app-sonar-scanner

go 1.12

require (
	github.com/GeertJohan/fgt v0.0.0-20160120143236-262f7b11eec0 // indirect
	github.com/antham/gommit v2.2.0+incompatible // indirect
	github.com/beevik/etree v1.1.0
	github.com/bxcodec/faker v2.0.1+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/gophercloud/gophercloud v0.1.0 // indirect
	github.com/jenkins-x/jx v0.0.0-20190717220822-c44dad43dfd4
	github.com/magiconair/properties v1.8.0
	github.com/pkg/errors v0.8.1
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.3.0
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a // indirect
	google.golang.org/appengine v1.5.0 // indirect
	k8s.io/api v0.0.0-20190718062839-c8a0b81cb10e // indirect
	k8s.io/apimachinery v0.0.0-20190717022731-0bb8574e0887
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/metrics v0.0.0-20190717024554-9e8c310e437b // indirect
)

replace github.com/heptio/sonobuoy => github.com/jenkins-x/sonobuoy v0.11.7-0.20190318120422-253758214767

replace k8s.io/api => k8s.io/api v0.0.0-20181128191700-6db15a15d2d3

replace k8s.io/metrics => k8s.io/metrics v0.0.0-20181128195641-3954d62a524d

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190122181752-bebe27e40fb7

replace k8s.io/client-go => k8s.io/client-go v2.0.0-alpha.0.0.20190115164855-701b91367003+incompatible

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20181128195303-1f84094d7e8e

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

replace github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v21.1.0+incompatible

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v10.15.5+incompatible

replace github.com/banzaicloud/bank-vaults => github.com/banzaicloud/bank-vaults v0.0.0-20190508130850-5673d28c46bd
