package util_test

import (
	"testing"

	. "github.com/Twister915/vis.go/pkg/util"
)

func BenchmarkCreateNDTs_3d(b *testing.B) {
	for n := 0; n < b.N; n++ {
		CreateNDTs(float64(1.0), 100, 100, 100)
	}
}

func BenchmarkCreate3DFloats(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Create3DFloats(100, 100, 100)
	}
}

func BenchmarkCreateNDTs_2d(b *testing.B) {
	for n := 0; n < b.N; n++ {
		CreateNDTs(float64(1.0), 100, 100)
	}
}

func BenchmarkCreate2DFloats(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Create2DFloats(100, 100)
	}
}

func Create3DFloats(x, y, z int) [][][]float64 {
	backing := make([]float64, MultiplyAll(x, y, z))

	n := 0
	out := make([][][]float64, x)
	for iX := range out {
		ys := make([][]float64, y)
		out[iX] = ys
		for iY := range ys {
			ys[iY] = backing[n : n+z]
			n += z
		}
	}

	return out
}
