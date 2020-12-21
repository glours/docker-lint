package internal

import (
	"fmt"
	"strings"
)

var (
	// Version is the version tag of the docker lint binary, set at build time
	Version = "unknown"
	// GitCommit is the commit of the docker lint binary, set at build time
	GitCommit = "unknown"
)

// FullVersion return plugin version, git commit and the provider cli version
func FullVersion() (string, error) {
	res := []string{
		fmt.Sprintf("Version:    %s", Version),
		fmt.Sprintf("Git commit: %s", GitCommit),
	}

	return strings.Join(res, "\n"), nil
}
