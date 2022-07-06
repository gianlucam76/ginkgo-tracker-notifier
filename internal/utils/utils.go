package utils

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2" // nolint: golint,stylecheck // ginkgo pattern
)

var logEnabled = false

func Init(doLogs bool) {
	logEnabled = doLogs
}

// Byf is a simple wrapper around By.
func Byf(format string, a ...interface{}) {
	if logEnabled {
		By(fmt.Sprintf(format, a...))
	}
}
