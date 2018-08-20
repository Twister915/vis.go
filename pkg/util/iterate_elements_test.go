package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Twister915/vis.go/pkg/util"
)

var _ = Describe("IterateElements", func() {
	It("iterates through all elements of a slice of ints", func() {
		slice := []int{1, 2, 3, 4}

		i := 0
		IterateElements(slice, func(idx []int, val interface{}) {
			Expect(i).ShouldNot(BeNumerically(">=", len(slice)))
			Expect(i).ShouldNot(BeNumerically("<", 0))

			Expect(idx).Should(HaveLen(1))
			Expect(idx[0]).Should(Equal(i))

			Expect(val).Should(BeAssignableToTypeOf(int(1)))
			vI := val.(int)

			Expect(vI).Should(Equal(slice[i]))
			i++
		})
	})

	It("iterates through all elements of a 2d slice of ints", func() {
		slice := [][]int{{1, 2, 3, 4}, {5, 6, 7, 8}, {9, 10, 11, 12}, {13, 14, 15, 16}}

		i, j, c := 0, 0, 0
		IterateElements(slice, func(idx []int, val interface{}) {
			Expect(i).ShouldNot(BeNumerically(">=", len(slice)))
			Expect(i).ShouldNot(BeNumerically("<", 0))
			Expect(j).ShouldNot(BeNumerically(">=", len(slice[i])))
			Expect(j).ShouldNot(BeNumerically("<", 0))

			Expect(idx).Should(HaveLen(2))
			Expect(idx[0]).Should(Equal(i))
			Expect(idx[1]).Should(Equal(j))

			Expect(val).Should(BeAssignableToTypeOf(int(1)))

			vI := val.(int)
			Expect(vI).Should(Equal(slice[i][j]))
			j++
			if j == len(slice[i]) {
				i++
				if i < len(slice) {
					j = 0
				}
			}

			c++
		})

		Expect(i).Should(Equal(4))
		Expect(j).Should(Equal(4))
		Expect(c).Should(Equal(16))
	})

	It("iterates through all elements of a 3d slice of ints", func() {
		slice := [][][]int{
			{{1, 2, 3, 4}, {5, 6, 7, 8}, {9, 10, 11, 12}, {13, 14, 15, 16}},
			{{17, 18, 19, 20}, {21, 22, 23, 24}, {25, 26, 27, 28}, {29, 30, 31, 32}},
			{{33, 34, 35, 36}, {37, 38, 39, 40}, {41, 42, 43, 44}, {45, 46, 47, 48}},
			{{49, 50, 51, 52}, {53, 54, 55, 56}, {57, 58, 59, 60}, {61, 62, 63, 64}},
		}

		i, j, k, c := 0, 0, 0, 0
		IterateElements(slice, func(idx []int, val interface{}) {
			Expect(i).ShouldNot(BeNumerically(">=", len(slice)))
			Expect(i).ShouldNot(BeNumerically("<", 0))
			Expect(j).ShouldNot(BeNumerically(">=", len(slice[i])))
			Expect(j).ShouldNot(BeNumerically("<", 0))
			Expect(k).ShouldNot(BeNumerically(">=", len(slice[i][j])))
			Expect(k).ShouldNot(BeNumerically("<", 0))

			Expect(idx).Should(HaveLen(3))
			Expect(idx[0]).Should(Equal(i))
			Expect(idx[1]).Should(Equal(j))
			Expect(idx[2]).Should(Equal(k))

			Expect(val).Should(BeAssignableToTypeOf(int(1)))

			vI := val.(int)
			Expect(vI).Should(Equal(slice[i][j][k]))

			k++
			if k == len(slice[i][j]) {
				k = 0
				j++
				if j == len(slice[i]) {
					j = 0
					i++
				}
			}

			c++
		})

		Expect(c).Should(Equal(64))
	})
})
