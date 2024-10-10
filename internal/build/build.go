package build

import "fmt"

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

const NA = "N/A"

func PrintGlobalVars() {
	if buildVersion == "" {
		buildVersion = NA
	}

	if buildDate == "" {
		buildDate = NA
	}

	if buildCommit == "" {
		buildCommit = NA
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
