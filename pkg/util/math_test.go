package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Twister915/vis.go/pkg/test_utils"
	. "github.com/Twister915/vis.go/pkg/util"
)

var _ = Describe("Math", func() {
	Describe("MultiplyAll", func() {
		It("should multiply all input numbers", func() {
			Expect(MultiplyAll(2, 3, 4, 5)).Should(Equal(2 * 3 * 4 * 5))
		})

		It("should not fail with 0 inputs", func() {
			Expect(MultiplyAll()).Should(Equal(0))
		})
	})

	Describe("Mean", func() {
		It("should average values", func() {
			Expect(Mean(1, 2, 3)).Should(BeEquivalentTo((1.0 + 2.0 + 3.0) / 3.0))
		})

		It("should not fail with 0 inputs", func() {
			Expect(Mean()).Should(BeNaN())
		})
	})
})
