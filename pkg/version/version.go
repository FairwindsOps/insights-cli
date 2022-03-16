// version stores the Insights CLI version, making it accessible by other
// packages.
package version

var version string = "unknown"
var commit string = "unknown"

func String() string {
	return "Version: " + version + " Commit: " + commit
}

func GetVersion() string {
	return version
}

func GetCommit() string {
	return commit
}
