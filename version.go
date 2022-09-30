package roamer

import "github.com/hashicorp/go-version"

const versionString = "0.1.1"

// GetVersionString returns the version string associated with this version of roamer.
func GetVersionString() string {
	return versionString
}

func getVersion() *version.Version {
	v, err := version.NewVersion(GetVersionString())
	if err != nil {
		panic(err)
	}
	return v
}
