package version

import (
	"fmt"
	"runtime"
)

// These variables are set at build time using ldflags
var (
	// Version is the semantic version (e.g., v1.0.0)
	Version = "dev"
	// Commit is the git commit hash
	Commit = "unknown"
	// Date is the build date
	Date = "unknown"
)

// Info returns formatted version information
func Info() string {
	return fmt.Sprintf("mdschema %s (%s) built on %s with %s",
		Version, Commit, Date, runtime.Version())
}

// Short returns just the version string
func Short() string {
	return Version
}
