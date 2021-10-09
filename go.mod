module github.com/spotahome/redis-operator

go 1.13

require (
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/prometheus/client_golang v1.1.0
	github.com/sirupsen/logrus v1.2.0
	github.com/spotahome/kooper v0.7.0
	github.com/stretchr/testify v1.4.0
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	k8s.io/api v0.0.0-20191004102349-159aefb8556b
	k8s.io/apiextensions-apiserver v0.0.0-20191114015135-f299f23b335b
	k8s.io/apimachinery v0.0.0-20191004074956-c5d2f014d689
	k8s.io/client-go v11.0.1-0.20191029005444-8e4128053008+incompatible
)
