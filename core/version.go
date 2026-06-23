package core

var AppVersion string

func init() {
	if AppVersion == "" {
		AppVersion = "dev"
	}
}
