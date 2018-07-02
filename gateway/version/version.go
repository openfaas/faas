package version

var (
	//Version release version of the provider
	Version string
	//GitCommit SHA of the last git commit
	GitCommitSHA string
	//GitCommit message of the last commit
	GitCommitMessage string
	//DevVersion string for the development version
	DevVersion = "dev"
)

//BuildVersion returns current version of the provider
func BuildVersion() string {
	if len(Version) == 0 {
		return DevVersion
	}
	return Version
}
