package util

type CombineFunc func(in ...float64) float64

func (c CombineFunc) Collapse(in [][]float64, out []float64) {
	for i, values := range in {
		out[i] = c(values...)
	}
}

func Transpose(in [][]float64) (out [][]float64) {
	out = Create2DFloats(len(in[0]), len(in))
	for i, vs := range in {
		for n, v := range vs {
			out[n][i] = v
		}
	}

	return
}
