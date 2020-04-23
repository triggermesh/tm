module github.com/triggermesh/tm

go 1.13

require (
	contrib.go.opencensus.io/exporter/ocagent v0.6.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.9 // indirect
	github.com/aws/aws-lambda-go v1.15.0
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/frankban/quicktest v1.7.2 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/knative/build v0.7.0
	github.com/knative/pkg v0.0.0-20190624141606-d82505e6c5b4
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/mholt/archiver v2.1.0+incompatible
	github.com/nwaples/rardecode v1.0.0 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/openzipkin/zipkin-go v0.2.2 // indirect
	github.com/pierrec/lz4 v2.4.0+incompatible // indirect
	github.com/prometheus/client_golang v1.1.0 // indirect
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/prometheus/common v0.7.0 // indirect
	github.com/prometheus/procfs v0.0.5 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/tektoncd/pipeline v0.11.2
	github.com/tektoncd/triggers v0.1.0
	github.com/tidwall/gjson v1.3.2 // indirect
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413
	google.golang.org/genproto v0.0.0-20191028173616-919d9bdd9fe6 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.2.5
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.0
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a // indirect
	k8s.io/utils v0.0.0-20191030222137-2b95a09bc58d // indirect
	knative.dev/eventing v0.12.0
	knative.dev/eventing-contrib v0.11.2
	knative.dev/pkg v0.0.0-20200207155214-fef852970f43
	knative.dev/serving v0.12.0
	sigs.k8s.io/controller-runtime v0.4.0 // indirect
)

replace (
	knative.dev/pkg => knative.dev/pkg v0.0.0-20200122022923-4e81bc3c320f
	k8s.io/api => k8s.io/api v0.0.0-20191016110246-af539daaa43a
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191016113439-b64f2075a530
	k8s.io/apimachinery => k8s.io/apimachinery v0.15.10-beta.0
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191016111841-d20af8c7efc5
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191016113937-7693ce2cae74
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191016110837-54936ba21026
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191016115248-b061d4666016
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20191016115051-4323e76404b0
	k8s.io/code-generator => k8s.io/code-generator v0.15.10-beta.0
	k8s.io/component-base => k8s.io/component-base v0.0.0-20191016111234-b8c37ee0c266
	k8s.io/cri-api => k8s.io/cri-api v0.15.10-beta.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20191016115443-72c16c0ea390
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191016112329-27bff66d0b7c
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191016114902-c7514f1b89da
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191016114328-7650d5e6588e
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191016114710-682e84547325
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20191016114520-100045381629
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191016115707-22244e5b01eb
	k8s.io/metrics => k8s.io/metrics v0.0.0-20191016113728-f445c7b35c1c
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191016112728-ceb381866e80
)
