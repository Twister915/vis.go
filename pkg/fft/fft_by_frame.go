package fft

//#include "fftw3.h"
import "C"

import (
	"math"
	"runtime"
	"time"
	"unsafe"

	"github.com/Twister915/vis.go/pkg/audio"
	"github.com/rs/zerolog/log"
)

// given some input audio, read blocks of data
func NewFFTByFrame(input audio.Input, windowSize time.Duration, windowMove time.Duration, window WindowingFunction, powTwo bool) *FFTByFrame {
	out := &FFTByFrame{
		input:      input,
		windowSize: windowSize,
		windowMove: windowMove,
		window:     window,
		powTwo:     powTwo,
	}

	out.initBuffer()
	out.initPlan()
	return out
}

type FFTByFrame struct {
	input audio.Input

	windowSize time.Duration
	windowMove time.Duration
	window     WindowingFunction
	powTwo     bool

	plan C.fftw_plan

	at time.Duration

	frameBuffer  FFTWDoubles2D
	resultBuffer FFTWComplexes2D

	windowPrecomputed []float64
}

func (f *FFTByFrame) initBuffer() {
	channels := f.input.Channels()
	desiredSamples := int(f.windowSize / f.input.Timebase())
	if f.powTwo {
		desiredSamples = NextPower2(desiredSamples)
	}
	f.frameBuffer = Alloc2DDoubles(desiredSamples, channels)
	f.resultBuffer = Alloc2DComplexes(channels, (desiredSamples/2)+1)

	a := CheckAlignment(f.frameBuffer, f.resultBuffer)
	log.Info().Bool("isAligned", a).Msg("alignment of frame & result buffer")

	if !a {
		panic("did not produce aligned arrays")
	}

	f.windowPrecomputed = make([]float64, desiredSamples)
	N := float64(desiredSamples)
	for i := range f.windowPrecomputed {
		f.windowPrecomputed[i] = f.window(float64(i), N)
	}
}

func (f *FFTByFrame) initPlan() {
	f.plan = fft_plan(
		time.Second*2, runtime.NumCPU(),
		len(f.frameBuffer), f.input.Channels(),

		unsafe.Pointer(&f.frameBuffer[0][0]), f.input.Channels(), 1,
		unsafe.Pointer(&f.resultBuffer[0][0]), 1, len(f.resultBuffer[0]),
		C.FFTW_ESTIMATE|C.FFTW_DESTROY_INPUT,
	)
}
func (f *FFTByFrame) HasNext() bool {
	return f.input.Has(len(f.frameBuffer))
}

// the argument passed is a destination
func (f *FFTByFrame) Compute(fftData [][]float64) (err error) {
	// read samples to fill the "frame buffer" (buffer which contains sample data for this frame)
	_, err = f.input.ReadSamples(f.frameBuffer)
	if err != nil {
		return
	}

	// now seek the file backwards so that we're ready to compute the next frame on the next call
	if err = f.input.Seek(int(f.windowMove/f.input.Timebase()) - len(f.frameBuffer)); err != nil {
		return
	}

	// now go through the frame buffer, and apply the windowing function to the values
	for i, frame := range f.frameBuffer {
		for chI, value := range frame {
			frame[chI] = f.windowPrecomputed[i] * value
		}
	}

	// finally, now that the data in f.frameBuffer is ready, perform the FFT on it (this plan will dump result
	// to f.resultBuffer)
	C.fftw_execute(f.plan)

	// read data from the resultBuffer (which is [][]complex128 and needs to become [][]float64 stored in fftData param)
	for cN, ch := range f.resultBuffer {
		for i, v := range ch {
			fftData[i][cN] = math.Hypot(real(v), imag(v))
		}
	}

	return
}

func (f *FFTByFrame) Close() error {
	defer f.resultBuffer.Free()
	defer f.frameBuffer.Free()

	func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		planMutex.Lock()
		defer planMutex.Unlock()

		C.fftw_destroy_plan(f.plan)
	}()

	return f.input.Close()
}

func (f *FFTByFrame) ComputeAll() (all [][][]float64, err error) {
	frames := int(f.input.Length() / f.windowMove)
	allN := make([]float64, frames*f.NumberOutputFrequencies()*f.input.Channels())
	all = make([][][]float64, frames)

	v := 0
	for frame := range all {
		all[frame] = make([][]float64, f.NumberOutputFrequencies())
		for bin := range all[frame] {
			end := v + f.input.Channels()
			all[frame][bin] = allN[v:end]
			v = end
		}
	}

	i := 0
	for f.HasNext() {
		if err = f.Compute(all[i]); err != nil {
			return
		}

		i++
	}

	return
}

// for each window, this is the number of frequencies we can detect (this is the sample rate, divided by two, plus one)
func (f *FFTByFrame) NumberOutputFrequencies() int {
	return len(f.resultBuffer[0])
}

func NextPower2(i int) int {
	return int(math.Pow(2, math.Ceil(math.Log2(float64(i)))))
}
