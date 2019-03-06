package credential

type GitCreds struct {
	Namespace string
	// Host string
	Key string
}

// RegistryCreds contains docker registry credentials
type RegistryCreds struct {
	Name      string
	Namespace string
	Host      string
	Username  string
	Password  string
	Pull      bool
	Push      bool
}
