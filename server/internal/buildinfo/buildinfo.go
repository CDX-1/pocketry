package buildinfo

import "fmt"

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func VersionString() string {
	return fmt.Sprintf("Pocketry Server %s (commit: %s - %s)", Version, Commit, Date)
}