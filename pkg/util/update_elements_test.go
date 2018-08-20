package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Twister915/vis.go/pkg/util"
)

var _ = Describe("UpdateElements", func() {
	It("updates a 1d slice of ints", func() {
		slice := []int{1, 2, 3}

		i := 0
		UpdateElements(slice, func(idx []int, v interface{}) interface{} {
			Expect(i).Should(BeNumerically(">=", 0))
			Expect(i).Should(BeNumerically("<", len(slice)))

			Expect(idx).Should(HaveLen(1))
			Expect(idx[0]).Should(Equal(i))

			Expect(v).Should(BeAssignableToTypeOf(int(0)))
			Expect(v).Should(Equal(slice[i]))

			out := v.(int) * 2
			i++

			return out
		})

		for i := range slice {
			Expect(slice[i]).Should(Equal((i + 1) * 2))
		}
	})

	It("updates a 2d slice of ints", func() {
		slice := [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}

		i := 0
		j := 0
		c := 0
		UpdateElements(slice, func(idx []int, v interface{}) interface{} {
			Expect(i).Should(BeNumerically(">=", 0))
			Expect(i).Should(BeNumerically("<", len(slice)))
			Expect(j).Should(BeNumerically(">=", 0))
			Expect(j).Should(BeNumerically("<", len(slice[i])))

			Expect(idx).Should(HaveLen(2))
			Expect(idx[0]).Should(Equal(i))
			Expect(idx[1]).Should(Equal(j))

			Expect(v).Should(BeAssignableToTypeOf(int(0)))
			Expect(v).Should(Equal(slice[i][j]))

			out := v.(int) * 2

			j++
			if j == len(slice[i]) {
				i++
				j = 0
			}

			c++

			return out
		})

		Expect(c).Should(Equal(9))
		c = 0
		for i := range slice {
			for j := range slice {
				Expect(slice[i][j]).Should(Equal((c + 1) * 2))
				c++
			}
		}
	})
})
