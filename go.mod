module github.com/triggermesh/tm

go 1.13

require (
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/googleapis/gnostic v0.4.2 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/nwaples/rardecode v1.1.0 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/afero v1.3.1
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.5.1
	github.com/tektoncd/pipeline v0.14.1
	github.com/tektoncd/triggers v0.6.1
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.5
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/klog/v2 v2.3.0 // indirect
	k8s.io/kube-openapi v0.0.0-20200615155156-dffdd1682719 // indirect
	k8s.io/utils v0.0.0-20200619165400-6e3d28b6ed19 // indirect
	knative.dev/eventing v0.16.0
	knative.dev/eventing-contrib v0.16.0
	knative.dev/networking v0.0.0-20200707203944-725ec013d8a2 // indirect
	knative.dev/pkg v0.0.0-20200713031612-b09a159e12c9
	knative.dev/serving v0.16.0
)

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0
	k8s.io/api => k8s.io/api v0.17.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.6
	k8s.io/apiserver => k8s.io/apiserver v0.17.6
	k8s.io/client-go => k8s.io/client-go v0.17.6
	k8s.io/code-generator => k8s.io/code-generator v0.17.6
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29
)
