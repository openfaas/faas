package main

var (
	//Version release version of the watchdog
	Version string
	//GitCommit SHA of the last git commit
	GitCommit string
	//DevVerison string for the development version
	DevVerison = "dev"
)

//BuildVersion returns current version of watchdog
func BuildVersion() string {
	if len(Version) == 0 {
		return DevVerison
	}
	return Version
}
