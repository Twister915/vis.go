package bin

import (
	"math"

	"github.com/Twister915/vis.go/pkg/util"
)

type CachedBinsSpec struct {
	Bins       []int
	FreqFactor float64
}

func PrecomputeBinSpec(N, sampleRate, bins int, fmin, fmax, gamma float64) (out *CachedBinsSpec) {
	//B_i = ((f_i / f_max) ** (1 / gamma)) * B_max
	// I redefine gamma to be the entire (1 / gamma) term in the above
	// gamma controls the level of detail, where higher values hide more detail
	gamma = 1.0 / gamma

	out = new(CachedBinsSpec)
	out.Bins = make([]int, bins+1)
	out.FreqFactor = float64(sampleRate+2) / float64((N-1)*2)

	for n := 0; n < N; n++ {
		frequency := out.FreqFactor * float64(n)

		if frequency < fmin {
			continue
		}

		bin := int(math.Pow(float64(frequency-fmin)/float64(fmax-fmin), gamma) * float64(bins))
		if bin < 0 {
			continue
		}

		if bin >= bins {
			out.Bins[bins] = n
			break
		}

		if out.Bins[bin] == 0 || out.Bins[bin] > n {
			out.Bins[bin] = n
		}
	}

	return
}

func (c *CachedBinsSpec) Bin(data [][]float64, to [][]float64) {
	nan := math.NaN()
	for i := range to {
		for n := range to[i] {
			to[i][n] = nan
		}
	}

	binI := 0
	for n, fftBin := range data {
		if n < c.Bins[binI] {
			continue
		}

		if n >= c.Bins[binI+1] {
			binI++
		}

		if binI >= (len(c.Bins) - 1) {
			break
		}

		for c, value := range fftBin {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				continue
			}

			if math.IsNaN(to[c][binI]) {
				to[c][binI] = 0
			}

			to[c][binI] += value
		}
	}
}

func Bin(data [][]float64, sampleRate, bins int, fmin, fmax, gamma float64, to [][]float64) (out [][]float64) {
	// create a slice of floats for the outputs
	channels := len(data[0])
	if to != nil {
		out = to
		if len(to) != channels || channels == 0 || len(to[0]) != bins {
			panic("bad out array passed")
		}

		nan := math.NaN()
		for i := range out {
			for n := range out[i] {
				out[i][n] = nan
			}
		}
	} else {
		// WARN: this allocates a new array for each binning operation
		out = util.Create2DFloats(channels, bins)
	}

	PrecomputeBinSpec(len(data), sampleRate, bins, fmin, fmax, gamma).Bin(data, out)
	return
}

// this combines channel values, which are currently separate, into a single value
func CombineChannelsAvg(in [][]float64, to []float64) (out []float64) {
	cs := float64(len(in))

	if to != nil {
		out = to
		if len(out) != len(in[0]) {
			panic("bad to slice passed, bad length")
		}
	} else {
		out = make([]float64, len(in[0]))
	}

	for i, val := range in[0] {
		out[i] = float64(val)
	}

	for _, ch := range in[1:] {
		for i, val := range ch {
			out[i] += float64(val)
		}
	}

	for i := range out {
		out[i] /= cs
	}

	return
}

// converts all input values to dB
func DBConversion(in []float64) {
	for i, v := range in {
		in[i] = math.Log10(v) * 10
	}
}

func DBConversionCs(in [][]float64) {
	for _, v := range in {
		DBConversion(v)
	}
}
