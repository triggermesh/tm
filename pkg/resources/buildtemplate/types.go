package buildtemplate

// Buildtemplate contains information about knative buildtemplate definition
type Buildtemplate struct {
	Name           string
	Namespace      string
	File           string
	RegistrySecret string
}
