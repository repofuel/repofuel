package version

var (
	// Version is populated at compile time by govvv from `./VERSION`.
	Version string

	// GitCommit is populated at compile time by govvv.
	// It will be obtained by `git rev-parse --short HEAD`.
	GitCommit string

	// GitState is populated at compile time by govvv.
	GitState string

	// GitSummary is populated at compile time by govvv.
	// It will be obtained by `git describe --tags --dirty --always`.
	GitSummary string

	// BuildDate is populated at compile time by govvv.
	BuildDate string

	// GitBranch is populated at compile time by govvv.
	// It will be obtained by `git symbolic-ref -q --short HEAD`.
	GitBranch string
)
