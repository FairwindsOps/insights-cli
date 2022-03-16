// Package version stores the Insights CLI version, making it accessible by other
// packages.
package version

var version string = "unknown"
var commit string = "unknown"

// String returns the version and git commit, as might be used as output for a
// `version` command-line flag.
func String() string {
	return "Version: " + version + " Commit: " + commit
}

// GetVersion returns the version unexported variable.
func GetVersion() string {
	return version
}

// GetCommit returns the commit unexported variable.
func GetCommit() string {
	return commit
}
