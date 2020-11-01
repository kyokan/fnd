package version

import "fmt"

var GitCommit string
var GitTag string
var UserAgent string

func init() {
	UserAgent = fmt.Sprintf("fnd/%s+%s", GitTag, GitCommit)
}
