package fft

//#include "fftw3.h"
import "C"

import (
	"runtime"
	"time"
	"unsafe"

	"github.com/Twister915/vis.go/pkg/audio"
	"github.com/Twister915/vis.go/pkg/util"
	"github.com/rs/zerolog/log"
)

func NewFFTComputeAll(input audio.Input, windowSize time.Duration, windowMove time.Duration, window WindowingFunction, powTwo bool) *FFTComputeAll {
	out := &FFTComputeAll{
		input: input,

		windowSize: windowSize,
		windowMove: windowMove,
		window:     window,
		powTwo:     powTwo,
	}

	out.initSizes()
	return out
}

type FFTComputeAll struct {
	input audio.Input

	windowSize time.Duration
	windowMove time.Duration
	window     WindowingFunction
	powTwo     bool

	frameCount      int
	samplesPerFrame int
}

func (f *FFTComputeAll) initSizes() {
	frameSamples := int(f.windowSize / f.input.Timebase())
	if f.powTwo {
		frameSamples = NextPower2(frameSamples)
		f.windowSize = time.Duration(frameSamples) * f.input.Timebase()
		if f.windowMove < f.windowSize {
			f.windowMove = f.windowSize
		}
	}

	f.samplesPerFrame = frameSamples
	f.frameCount = int((f.input.Length() - f.windowSize) / f.windowMove)
}

func (f *FFTComputeAll) initPlan(frameBuffer [][][]float64, resultBuffer [][][]complex128) C.fftw_plan {
	return fft_plan(
		time.Second,
		runtime.NumCPU(), f.samplesPerFrame, f.input.Channels()*f.frameCount,
		unsafe.Pointer(&frameBuffer[0][0][0]), 1, f.samplesPerFrame,
		unsafe.Pointer(&resultBuffer[0][0][0]), 1, f.samplesPerFrame/2,
		C.FFTW_ESTIMATE|C.FFTW_DESTROY_INPUT,
	)
}

func (f *FFTComputeAll) closePlan(plan C.fftw_plan) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	planMutex.Lock()
	defer planMutex.Unlock()

	C.fftw_destroy_plan(plan)
}

func (f *FFTComputeAll) readAudioToFrames(frameBuffer [][][]float64) (err error) {
	// arrange the audio data into the right shape
	back := int((f.windowSize - f.windowMove) / f.input.Timebase())
	for i := range frameBuffer {
		if _, err = f.input.ReadSamplesDir(frameBuffer[i], audio.ReadChannelBySample); err != nil {
			return
		}

		for _, ch := range frameBuffer[i] {
			sum := 0.0
			for i, v := range ch {
				v *= f.window(float64(i), float64(f.samplesPerFrame))
				ch[i] = v
				sum += v
			}

			mean := sum / float64(len(ch))

			for i := range ch {
				ch[i] = ch[i] - mean
			}
		}

		if err = f.input.Seek(-back); err != nil {
			return
		}
	}

	return
}

func (f *FFTComputeAll) ComputeAll() (all [][][]float64, err error) {
	channels := f.input.Channels()
	log.Info().
		Int("channels", channels).
		Int("frames", f.frameCount).
		Int("samplesPerFrame", f.samplesPerFrame).
		Int("samplesMove", int(f.windowMove/f.input.Timebase())).
		Msg("start compute all")
	// the data is organized like this
	// each [][]float64 contains all the channels & samples for one FFT frame
	// each []float64 is a channel, with a series of samples for that channel
	// each float64 is a sample
	// the earliest data is always at the lowest index
	// if you did framesBuffer[10][0][99], you would get the 100th sample on the 1st channel for the 10th frame
	// the length of the [][][]float64 is the number of frames
	// the length of each [][]float64 is the number of channels
	// the length of each []float64 is the samples per frame
	frameBuffer := Alloc3DDoubles(f.frameCount, channels, f.samplesPerFrame)
	defer frameBuffer.Free()
	// this data is organized like this:
	// each [][]complex128 is a frame of FFT output
	// each []complex128 is a channel of FFT output
	// each complex128 is a value from the FFT
	// the earliest data is always at the lowest index
	// also, the size of the FFT output is sampleRate / 2
	// if you did resultsBuffer[11][1][199], you would get the 200th bin from the 2nd channel of the 12th frame
	resultBuffer := Alloc3DComplexes(f.frameCount, channels, f.samplesPerFrame/2)
	defer resultBuffer.Free()
	if err = f.readAudioToFrames(frameBuffer); err != nil {
		return
	}

	log.Info().Msg("init fft planning...")
	plan := f.initPlan(frameBuffer, resultBuffer)
	log.Info().Msg("complete fft planning")
	defer f.closePlan(plan)

	log.Info().
		Int("ffts", f.input.Channels()*f.frameCount).
		Int("size", f.samplesPerFrame).
		Msg("executing plan")

	C.fftw_execute(plan)

	log.Info().Msg("finished ffts, doing GC")

	runtime.GC()

	log.Info().Msg("writing output data")
	// output data, organized by frames, then samples, then channels
	all = util.CreateNDFloat64(f.frameCount, f.samplesPerFrame/2, channels).([][][]float64)

	// bring all the data to the all slice
	for fI, frame := range resultBuffer {
		for chI, channel := range frame {
			for i, v := range channel {
				// notice the indices are flipped
				r, im := real(v), imag(v)
				all[fI][i][chI] = (r * r) + (im * im)
			}
		}
	}

	log.Info().Msg("computed all frames")
	return
}
