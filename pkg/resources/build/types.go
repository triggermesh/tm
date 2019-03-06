package build

// Build structure represents knative build object
type Build struct {
	Name          string
	Namespace     string
	Source        string
	Revision      string
	Step          string
	Command       []string
	Buildtemplate string
	Args          []string
	Image         string
}
