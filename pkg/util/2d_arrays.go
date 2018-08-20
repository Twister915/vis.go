package util

// allocates a giant array
func Create2DFloats(x, y int) [][]float64 {
	return Reshape2DFloats(x, y, make([]float64, x*y))
}

func Reshape2DFloats(x, y int, backing []float64) (out [][]float64) {
	if x*y != len(backing) {
		panic("invalid backing slice passed")
	}

	out = make([][]float64, x)

	for i := range out {
		n := y * i
		out[i] = backing[n : n+y]
	}

	return
}

func Create2DComplex(x, y int) [][]complex128 {
	return Reshape2DComplex(x, y, make([]complex128, x*y))

}

func Reshape2DComplex(x, y int, backing []complex128) (out [][]complex128) {
	if x*y != len(backing) {
		panic("invalid backing slice passed")
	}

	out = make([][]complex128, x)

	for i := range out {
		n := y * i
		out[i] = backing[n : n+y]
	}

	return
}
