module github.com/spotahome/redis-operator

go 1.16

require (
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/prometheus/client_golang v1.11.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spotahome/kooper v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.22.3
	k8s.io/apiextensions-apiserver v0.22.3
	k8s.io/apimachinery v0.22.3
	k8s.io/client-go v0.22.3
)

replace github.com/spotahome/kooper => github.com/yxxhero/kooper v0.8.1-0.20211030060712-a781cb073699

replace k8s.io/client-go => k8s.io/client-go v0.22.3
