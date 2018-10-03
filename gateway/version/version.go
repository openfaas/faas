package version

var (
	// Version release version of the provider
	Version string

	// GitCommitSHA is the Git SHA of the latest tag/release
	GitCommitSHA string

	// GitCommitMessage as read from the latest tag/release
	GitCommitMessage string

	// DevVersion string for the development version
	DevVersion = "dev"
)

// BuildVersion returns current version of the provider
func BuildVersion() string {
	if len(Version) == 0 {
		return DevVersion
	}
	return Version
}
