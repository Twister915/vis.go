package main

import (
	"math"
	"time"

	"github.com/Twister915/vis.go/pkg/audio"
	"github.com/Twister915/vis.go/pkg/bin"
	"github.com/Twister915/vis.go/pkg/fft"
	"github.com/Twister915/vis.go/pkg/util"
	"github.com/rs/zerolog/log"
)

type streamingFFT struct {
	Audio  audio.Input
	Window fft.WindowingFunction `json:"-"`

	FMin, FMax float64
	Gamma      float64
	Bins       int
	WindowSize time.Duration
	FrameRate  int
	PowTwo     bool

	SmoothingAlpha float64
	PercentileHigh float64
	PercentileLow  float64

	SmoothingPasses int
	SmoothingPoints int

	fftBuffer      [][]float64
	binBuffer      [][]float64
	combinedBuffer []float64

	hasLastBinned bool
	lastBinned    [][]float64

	EstimateFrameSize time.Duration
	EstimateStride    time.Duration

	rollingMu    float64
	rollingSigma float64
	nMuSigma     int

	fft            *fft.FFTByFrame
	precomputedBin *bin.CachedBinsSpec
}

type FFTResult struct {
	I    int
	Err  error
	Data [][]float64
}

func (f *streamingFFT) init() {
	log.Info().Interface("config", f).Msg("init streaming FFT")
	samplesPerFrame := int(f.WindowSize / f.Audio.Timebase())
	if f.PowTwo {
		samplesPerFrame = fft.NextPower2(samplesPerFrame)
		f.WindowSize = time.Duration(samplesPerFrame) * f.Audio.Timebase()
		log.Info().Str("dur", f.WindowSize.String()).Int("samplesPerFrame", samplesPerFrame).Msg("rounded to pow2")
	}

	channels := f.Audio.Channels()

	f.fftBuffer = util.Create2DFloats(samplesPerFrame, channels)
	f.binBuffer = util.Create2DFloats(channels, f.Bins)
	f.lastBinned = util.Create2DFloats(channels, f.Bins)
	f.combinedBuffer = make([]float64, f.Bins)
	f.fft = fft.NewFFTByFrame(f.Audio, f.WindowSize, time.Second/time.Duration(f.FrameRate), f.Window, false)
	f.precomputedBin = bin.PrecomputeBinSpec(f.fft.NumberOutputFrequencies(), f.Audio.SampleRate(), f.Bins, f.FMin, f.FMax, f.Gamma)
}

func (f *streamingFFT) StreamFFT(to chan<- FFTResult) {
	var err error
	var i int
	defer func() {
		defer close(to)

		if err != nil {
			to <- FFTResult{I: i, Err: err}
		}
	}()

	err = f.estimateDistribution()
	if err != nil {
		return
	}

	buf := util.CreateNDFloat64(cap(to)+2, f.Audio.Channels(), f.Bins).([][][]float64)
	nB := 0
	for f.fft.HasNext() {
		if err = f.computeFrameFFT(); err != nil {
			return
		}

		result := buf[nB]
		nB++
		if nB >= len(buf) {
			nB = 0
		}

		f.bin(result)
		bin.CombineChannelsAvg(result, f.combinedBuffer)
		f.addValuesToMuSigma(i, f.combinedBuffer)
		f.postBinProcessing(result)
		to <- FFTResult{I: i, Data: result}
		i++
	}

	log.Info().Float64("mu", f.rollingMu).Float64("sigma", f.rollingSigma).Msg("final sigma & µ")
}

func (f *streamingFFT) estimateDistribution() (err error) {
	sizeFrames := int(f.EstimateFrameSize / (time.Second / time.Duration(f.FrameRate)))
	strideSamples := int((f.EstimateStride - f.EstimateFrameSize) / f.Audio.Timebase())
	n := int(f.Audio.Length() / f.EstimateStride)

	log.Info().
		Int("n", n).
		Int("frames", sizeFrames).
		Int("strideSamples", strideSamples).
		Str("stride", f.EstimateStride.String()).
		Str("tb", f.Audio.Timebase().String()).
		Str("len", f.Audio.Length().String()).
		Str("size", f.EstimateFrameSize.String()).
		Int("ffts", n*sizeFrames*f.Audio.Channels()).
		Msg("estimating distribution")

	data := util.Create2DFloats(sizeFrames*n, f.Bins)

	if err = f.Audio.Reset(); err != nil {
		return
	}

	f.nMuSigma = 0
	for i := 0; i < n; i++ {
		for j := 0; j < sizeFrames; j++ {
			if err = f.computeFrameFFT(); err != nil {
				return
			}

			f.binCombined(data[i*sizeFrames+j])
			f.nMuSigma += len(data[i*sizeFrames+j])
		}

		f.Audio.Seek(strideSamples)
	}

	f.rollingMu, f.rollingSigma = analyzeDistribution(data)
	log.Info().Float64("mu", f.rollingMu).Float64("sigma", f.rollingSigma).Msg("estimated mu & sigma")
	err = f.Audio.Reset()
	return
}

func (f *streamingFFT) computeFrameFFT() error {
	return f.fft.Compute(f.fftBuffer)
}

func (f *streamingFFT) postBinProcessing(binnedCs [][]float64) {
	for _, binned := range binnedCs {
		normalizeExtrema(binned, f.rollingMu, f.rollingSigma, f.PercentileLow, f.PercentileHigh)
		savitskyGolaySmooth(binned, f.SmoothingPasses, f.SmoothingPoints)
	}
	f.exponentialSmoothing(binnedCs)
}

func (f *streamingFFT) bin(to [][]float64) {
	f.precomputedBin.Bin(f.fftBuffer, to)
	bin.DBConversionCs(to)
}

func (f *streamingFFT) binCombined(to []float64) {
	f.precomputedBin.Bin(f.fftBuffer, f.binBuffer)
	bin.CombineChannelsAvg(f.binBuffer, to)
	bin.DBConversion(to)
}

func (f *streamingFFT) exponentialSmoothing(in [][]float64) {
	defer func() {
		f.hasLastBinned = true
		for i := range in {
			copy(f.lastBinned[i], in[i])
		}
	}()

	if !f.hasLastBinned {
		return
	}

	ma := 1 - f.SmoothingAlpha
	for c, cVs := range in {
		for i, v := range cVs {
			old := f.lastBinned[c][i]
			if math.IsInf(v, 0) || math.IsNaN(v) || math.IsNaN(old) || math.IsInf(old, 0) {
				continue
			}

			cVs[i] = (f.SmoothingAlpha * v) + (ma * old)
		}
	}
}

func (f *streamingFFT) addValuesToMuSigma(i int, in []float64) {
	if f.isInEstimate(i) {
		return
	}

	for _, v := range in {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}

		f.rollingMu, f.rollingSigma = updateMuSigmaWithValue(f.rollingMu, f.rollingSigma, v, f.nMuSigma)
		f.nMuSigma++
	}
}

func (f *streamingFFT) isInEstimate(i int) bool {
	rate := time.Second / time.Duration(f.FrameRate)
	sizeFrames := int(f.EstimateFrameSize / rate)
	strideFrames := int(f.EstimateStride / rate)

	return i%strideFrames <= sizeFrames
}

func updateMuSigmaWithValue(mu, sigma, introduce float64, n int) (newMu, newSigma float64) {
	newMu = ((mu * float64(n)) + introduce) / float64(n+1)
	newSigma = math.Sqrt(((float64(n-1) * (sigma * sigma)) + ((introduce - newMu) * (introduce - mu))) / float64(n))

	//dm := newMu - mu
	//ds := newSigma - sigma
	//if dm > 0 || ds > 0 {
	//	log.Info().
	//		Float64("mu", newMu).
	//		Float64("sig", newSigma).
	//		Float64("∂mu", dm).
	//		Float64("∂sigma", ds).
	//		Int("n", n).
	//		Msg("updated mu and sigma")
	//}
	return
}
