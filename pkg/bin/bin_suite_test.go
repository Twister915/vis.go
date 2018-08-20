package bin_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bin Suite")
}
