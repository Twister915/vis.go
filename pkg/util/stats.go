package util

import "math"

func Normalize(in [][]float64) (out [][]float64, min, max float64) {
	min = math.Inf(1)
	max = math.Inf(-1)

	for _, r := range in {
		for _, v := range r {
			if min > v {
				min = v
			}

			if max < v {
				max = v
			}
		}
	}

	out = Create2DFloats(len(in), len(in[0]))

	maxMinusMin := max - min
	for n, r := range in {
		for i, v := range r {
			out[n][i] = (v - min) / maxMinusMin
		}
	}

	return
}
