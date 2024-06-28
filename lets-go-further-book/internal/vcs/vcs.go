package vcs

import "fmt"

var (
    version   string
    buildTime string
    gitCommit string
)

func Version() string {
    if version == "" && buildTime == "" && gitCommit == "" {
        return "unknown-no-vcs-info"
    }
    return fmt.Sprintf("version: %s, build time: %s, git commit: %s", version, buildTime, gitCommit)
}
