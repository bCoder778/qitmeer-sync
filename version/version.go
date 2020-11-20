package version

import (
	"fmt"
	"runtime"
)

var (
	// 初始化为 unknown，如果编译时没有传入这些值，则为 unknown
	GitCommitLog   = "unknown"
	GitStatus      = "unknown"
	BuildTime      = "unknown"
	BuildGoVersion = "unknown"
	Version        = "v0.1.0"
)

// 返回单行格式
func StringifySingleLine() string {
	return fmt.Sprintf("Qitmeer-sync Version=%s. GitCommitLog=%s. GitStatus=%s. BuildTime=%s. GoVersion=%s. runtime=%s/%s.",
		Version, GitCommitLog[0:10], GitStatus, BuildTime, BuildGoVersion, runtime.GOOS, runtime.GOARCH)
}

// 返回多行格式
func StringifyMultiLine() string {
	return fmt.Sprintf("Qitmeer-sync = %s\nGitCommit = %s\nGitStatus=%s\nBuildTime = %s\nGoVersion = %s\n",
		Version, GitCommitLog[0:10], GitStatus, BuildTime, BuildGoVersion)
}
