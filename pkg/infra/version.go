package infra

import (
	"fmt"
	"runtime"
)

const (
	programName = "tape++"
)

var (
	Version   string = "latest"
	CommitSHA string = "development build"
	BuiltTime string = "Sat Nov 11 15:00:00 2023"
)

// GetVersionInfo return version information
// TODO add commit hash, Built info
func GetVersionInfo() string {
	return fmt.Sprintf(
		"%s:\n Version: %s\n Go version: %s\n Git commit: %s\n Built: %s\n OS/Arch: %s\n",
		programName,
		Version,
		runtime.Version(),
		CommitSHA,
		BuiltTime,
		fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	)
}
