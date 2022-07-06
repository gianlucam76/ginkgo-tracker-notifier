package process_result_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProcessResult(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "ProcessResult Suite")
}
