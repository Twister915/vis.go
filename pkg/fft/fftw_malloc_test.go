package fft_test

import (
	"testing"

	"github.com/Twister915/vis.go/pkg/fft"
	"github.com/davecgh/go-spew/spew"
)

func TestAllocFFTWDoubles(t *testing.T) {
	doubles := fft.AllocFFTWDoubles(256)
	defer doubles.Free()

	t.Logf("ptr => %p", &doubles[0])

	for i := 0; i < len(doubles); i++ {
		doubles[i] = float64(i + 1)
	}

	t.Logf(spew.Sdump(doubles))
}

func TestAllocFFTWComplexes(t *testing.T) {
	complexes := fft.AllocFFTWComplexes(256)
	defer complexes.Free()

	t.Logf("ptr => %p", &complexes[0])

	for i := 0; i < len(complexes); i++ {
		complexes[i] = complex(float64(i+1), float64(2*i))
	}

	t.Logf(spew.Sdump(complexes))
}
