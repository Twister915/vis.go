package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Twister915/vis.go/pkg/test_utils"
	. "github.com/Twister915/vis.go/pkg/util"
)

var _ = Describe("Arrays", func() {
	Describe("straightforward construction", func() {
		It("creates a 2d contiguous slice", func() {
			const (
				x = 4
				y = 4
			)

			By("creating a slice of floats")
			{
				myArray := Create2DFloats(x, y)
				Expect(myArray).Should(BeContiguous())
				Expect(myArray).Should(HaveLen(x))
				Expect(myArray[0]).Should(HaveLen(y))
			}

			By("creating a slice of complex numbers")
			{
				myArray := Create2DComplex(x, y)
				Expect(myArray).Should(BeContiguous())
				Expect(myArray).Should(HaveLen(x))
				Expect(myArray[0]).Should(HaveLen(y))
			}
		})

		It("creates a 3d contiguous slice", func() {
			const (
				x = 4
				y = 4
				z = 4
			)

			By("creating a slice of floats")
			myArray := Create3DFloats(x, y, z)
			Expect(myArray).Should(BeContiguous())
			Expect(myArray).Should(HaveLen(x))
			Expect(myArray[0]).Should(HaveLen(y))
			Expect(myArray[0][0]).Should(HaveLen(z))
		})
	})

	Describe("reflection based construction", func() {
		It("creates a 2d contiguous slice", func() {
			const (
				x = 4
				y = 4
			)

			By("creating a slice of floats")
			{
				rawArray := CreateNDFloat64(x, y)
				Expect(rawArray).Should(BeAssignableToTypeOf([][]float64{}))

				myArray := rawArray.([][]float64)
				Expect(myArray).Should(BeContiguous())
				Expect(myArray).Should(HaveLen(x))
				Expect(myArray[0]).Should(HaveLen(y))
			}

			By("creating a slice of complex numbers")
			{
				rawArray := CreateNDComplex128(x, y)
				Expect(rawArray).Should(BeAssignableToTypeOf([][]complex128{}))

				myArray := rawArray.([][]complex128)
				Expect(myArray).Should(BeContiguous())
				Expect(myArray).Should(HaveLen(x))
				Expect(myArray[0]).Should(HaveLen(y))
			}

			By("creating a slice of int64s")
			{
				rawArray := CreateNDTs(int64(0), x, y)
				Expect(rawArray).Should(BeAssignableToTypeOf([][]int64{}))

				myArray := rawArray.([][]int64)
				Expect(myArray).Should(BeContiguous())
				Expect(myArray).Should(HaveLen(x))
				Expect(myArray[0]).Should(HaveLen(y))
			}
		})
	})
})
