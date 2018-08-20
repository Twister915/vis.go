package bin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"math"
	"time"

	. "github.com/Twister915/vis.go/pkg/bin"
	"github.com/Twister915/vis.go/pkg/util"
	"github.com/davecgh/go-spew/spew"
)

func decPlaces(i float64, n int) float64 {
	fac := math.Pow(10, float64(n))
	return float64(int(i*fac)) / fac
}

var _ = Describe("Binner", func() {
	const (
		sampleRate = 44100
		windowSize = time.Millisecond * 300
	)

	var exampleData, resultData [][]float64
	W := float64(windowSize) / float64(time.Second)
	N := int(float64(sampleRate/2)*W) + 1
	gamma := 2.0
	fmin := 40.0
	fmax := 10000.0
	bins := 24
	channels := 2

	BeforeEach(func() {
		exampleData = util.Create2DFloats(N, channels)
		for n, d := range exampleData {
			for i := range d {
				d[i] = float64(n)
			}
		}

		resultData = util.Create2DFloats(channels, bins)
	})

	toHz := func(i int) float64 {
		return float64((sampleRate+2)*i) / float64((N-1)*2)
	}

	binForI := func(i int) int {
		return int(math.Pow(float64(toHz(i)-fmin)/float64(fmax-fmin), 1/gamma) * float64(bins))
	}

	startAt := func(binI int) int {
		out := int(((math.Pow(float64(binI)/float64(bins), gamma) * (fmax - fmin)) + fmin) * W)
		for binForI(out) > binI {
			out--
		}

		for binForI(out) < binI {
			out++
		}

		return out
	}

	dataRange := func(binI int) (f, l int) {
		f = startAt(binI)
		l = startAt(binI+1) - 1
		return
	}

	expectedVal := func(binI int) (out float64) {
		f, l := dataRange(binI)
		out = 0
		for i := f; i <= l; i++ {
			out += float64(i)
		}
		return
	}

	It("computes Hz correctly", func() {
		hzPerBin := (float64(sampleRate + 2)) / (float64(N * 2))

		checkHz := func(i int) {
			hz := decPlaces(toHz(i), 5)
			hzC := decPlaces(hzPerBin*float64(i), 5)
			Expect(hz).Should(Equal(hzC), "bin %d from input space = %.02fHz for %d = N, %d = sr", i, toHz(i), N, sampleRate)
		}

		for i := 0; i < N; i++ {
			checkHz(i)
		}
	})

	It("computes bin range correctly", func() {
		Context("0th bin", func() {
			s, e := dataRange(0)
			Expect(s).To(Equal(13))
			Expect(e).To(Equal(17))
		})

		Context("1st bin", func() {
			s, _ := dataRange(1)
			Expect(s).To(Equal(18))
			//Expect(e).To(Equal(18))
		})
	})

	It("bins correctly", func() {
		Bin(exampleData, sampleRate, bins, fmin, fmax, gamma, resultData)
		spew.Fdump(GinkgoWriter, resultData)
		for _, c := range resultData {
			for i, v := range c {
				start, end := dataRange(i)
				fmt.Fprintf(GinkgoWriter, "bin for %d (%.02fHz) was %d\n", end, toHz(end), binForI(end))
				Expect(v).Should(Equal(expectedVal(i)), "result bin %d should be sum between %d -> %d (%.01fHz -> %.01fHz)", i, start, end, toHz(start), toHz(end))
			}
		}
	})
})
