module github.com/spotahome/kooper

// Dependencies we don't really need, except that kubernetes specifies them as v0.0.0 which confuses go.mod
//replace k8s.io/apiserver => k8s.io/apiserver kubernetes-1.15.6
//replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver kubernetes-1.15.6
//replace k8s.io/api => k8s.io/api kubernetes-1.15.6
//replace k8s.io/component-base => k8s.io/component-base kubernetes-1.15.6
//replace k8s.io/client-go => k8s.io/client-go kubernetes-1.15.6
//replace k8s.io/kube-scheduler => k8s.io/kube-scheduler kubernetes-1.15.6
//replace k8s.io/apimachinery => k8s.io/apimachinery kubernetes-1.15.6
//replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers kubernetes-1.15.6
//replace k8s.io/kubelet => k8s.io/kubelet kubernetes-1.15.6
//replace k8s.io/cloud-provider => k8s.io/cloud-provider kubernetes-1.15.6
//replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib kubernetes-1.15.6
//replace k8s.io/cli-runtime => k8s.io/cli-runtime kubernetes-1.15.6
//replace k8s.io/kube-aggregator => k8s.io/kube-aggregator kubernetes-1.15.6
//replace k8s.io/sample-apiserver => k8s.io/sample-apiserver kubernetes-1.15.6
//replace k8s.io/metrics => k8s.io/metrics kubernetes-1.15.6
//replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap kubernetes-1.15.6
//replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager kubernetes-1.15.6
//replace k8s.io/kube-proxy => k8s.io/kube-proxy kubernetes-1.15.6
//replace k8s.io/cri-api => k8s.io/cri-api kubernetes-1.15.6
//replace k8s.io/code-generator => k8s.io/code-generator kubernetes-1.15.6

replace k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191114102923-bf973bc1a46c

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191114105316-e8706470940d

replace k8s.io/api => k8s.io/api v0.0.0-20191114100237-2cd11237263f

replace k8s.io/component-base => k8s.io/component-base v0.0.0-20191114102239-843ff05e8ff4

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191114101336-8cba805ad12d

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191114111147-29226eb67741

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191004115701-31ade1b30762

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191114112557-fb8eac6d1d79

replace k8s.io/kubelet => k8s.io/kubelet v0.0.0-20191114110913-8a0729368279

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191114111940-b2efa58ca04c

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20191114112225-e438b10da852

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191114110057-22fabc8113ba

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191114103707-3917fe134eab

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191114104325-4dc280b03897

replace k8s.io/metrics => k8s.io/metrics v0.0.0-20191114105745-bf91bab17669

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20191114111701-466976f32df4

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191114111427-e269b4a0667c

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191114110636-5b9a03eee945

replace k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190817025403-3ae76f584e79

replace k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b

require (
	github.com/Pallinder/go-randomdata v0.0.0-20180329154440-dab270d296c6
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gophercloud/gophercloud v0.4.0 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/prometheus/client_golang v1.1.0
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/uber-go/atomic v1.4.0 // indirect
	github.com/uber/jaeger-client-go v2.19.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	go.uber.org/atomic v1.4.0 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/sync v0.0.0-20190227155943-e225da77a7e6
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.0.0
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v0.0.0
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kube-openapi v0.0.0-20190918143330-0270cf2f1c1d // indirect
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6 // indirect
)

go 1.13
