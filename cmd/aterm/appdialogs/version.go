package appdialogs

import (
	"github.com/theparanoids/aterm/cmd/aterm/config"
)

// PrintVersion prints a simple sentence that lists the current version and commit hash
// Note: this should only be called _after_ parsing CLI options
func PrintVersion() {
	printline("ATerm Version:", config.Version(), " Build Hash:", config.CommitHash())
}

// PrintExtendedVersion prints more version information: the go runtime and the build date
// Note: this should only be called _after_ parsing CLI options
func PrintExtendedVersion() {
	printline("Go runtime:", config.GoRuntime(), " Build Date:", config.BuildDate())
}
