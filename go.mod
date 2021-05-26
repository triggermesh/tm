module github.com/triggermesh/tm

go 1.13

require (
	github.com/Microsoft/go-winio v0.4.15 // indirect
	github.com/alecthomas/units v0.0.0-20201120081800-1786d5ef83d4 // indirect
	github.com/antlr/antlr4 v0.0.0-20201124223952-86610fd4b43e // indirect
	github.com/frankban/quicktest v1.10.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/golang/snappy v0.0.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20201106050909-4977a11b4351 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/nwaples/rardecode v1.1.0 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/rickb777/date v1.14.3 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/afero v1.4.1
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/tektoncd/pipeline v0.18.1
	github.com/tektoncd/triggers v0.10.0
	github.com/ulikunitz/xz v0.5.8 // indirect
	github.com/xanzy/ssh-agent v0.3.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.17.0 // indirect
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/eventing v0.23.1
	knative.dev/eventing-github v0.19.0
	knative.dev/pkg v0.0.0-20210526081028-980a33719a10
	knative.dev/serving v0.22.2
)

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0
	k8s.io/api => k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8
	k8s.io/apiserver => k8s.io/apiserver v0.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
	k8s.io/code-generator => k8s.io/code-generator v0.18.8
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29
)
