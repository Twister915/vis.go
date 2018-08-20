package main

import (
	"math"
	"os"
	"runtime"
	"time"
	"sync"

	"github.com/Twister915/vis.go/pkg/fft"
	"github.com/Twister915/vis.go/pkg/util"
	"github.com/Twister915/vis.go/pkg/wav"
	"github.com/hajimehoshi/oto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/mmap"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	runtime.LockOSThread()
	var wg sync.WaitGroup
	defer wg.Wait()

	shutdown := make(chan struct{})
	defer close(shutdown)

	wg.Add(2)
	go memoryProfileHandler(&wg, shutdown)
	go cpuProfileHandler(&wg, shutdown)

	if len(os.Args) == 1 {
		panic("must supply file to read")
	}

	initLog()

	window := NewWindow()
	if err := window.Init(1280, 720, os.Getenv("FS") == "true", "Visualizer - ..."); err != nil {
		panic(err)
	}

	go window.UpdateLoop(60)

	go func() {
		for _, fname := range os.Args[1:] {
			window.w.SetTitle("Visualizer - " + fname)
			doFFT(window, fname)
		}

		window.Close()
	}()

	window.DrawLoop()
}

func initLog() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func doFFT(window *window, fileName string) {
	m, err := mmap.Open(fileName)
	if err != nil {
		panic(err)
	}

	fftInput, err := wav.ReadWav(&util.MMapSeeker{M: m})
	if err != nil {
		panic(err)
	}

	streamer := &streamingFFT{
		Audio:             fftInput,
		Window:            fft.BlackmanNuttallWindow,
		FMin:              20,
		FMax:              18000,
		Bins:              40,
		Gamma:             2.2,
		WindowSize:        time.Millisecond * 165,
		FrameRate:         30,
		SmoothingAlpha:    0.45,
		PercentileHigh:    0.07,
		PercentileLow:     0.4,
		SmoothingPasses:   1,
		SmoothingPoints:   3,
		EstimateFrameSize: time.Millisecond * 800,
		EstimateStride:    time.Millisecond * 6500,
	}

	streamer.init()

	binsPerFrame := streamer.fft.NumberOutputFrequencies()
	log.Info().
		Int("sampleRate", fftInput.SampleRate()).
		Int("n", binsPerFrame).
		Int("bins", streamer.Bins).
		Float64("fmin", streamer.FMin).
		Float64("fmax", streamer.FMax).
		Float64("gamma", streamer.Gamma).
		Msg("params")

	for i := 0; i < streamer.Bins; i++ {
		start, end, startN, endN := dataRange(i, streamer.Bins, fftInput.SampleRate(), streamer.FMax, streamer.FMin, streamer.Gamma, streamer.WindowSize)
		log.Info().Msgf("bin[%d]: %.03fHz (%d) -> %.03fHz (%d) (%d = size)", i+1, start, startN, end, endN, IntAbs(endN-startN)+1)
	}

	to := make(chan FFTResult, streamer.FrameRate*10)
	go streamer.StreamFFT(to)

	audioPlay, err := wav.ReadWav(&util.MMapSeeker{M: m})
	if err != nil {
		panic(err)
	}

	p := &player{input: audioPlay}
	frameBuf := make([]byte, (audioPlay.SampleRate()/streamer.FrameRate)*audioPlay.Channels()*(audioPlay.BitDepth()/8))
	audioStreamer := p.Stream(audioPlay.BitDepth(), frameBuf)
	oPlayer, err := oto.NewPlayer(audioPlay.SampleRate(), audioPlay.Channels(), audioPlay.BitDepth()/8, len(frameBuf))
	if err != nil {
		panic(err)
	}

	playAudio := func() {
		if err = audioStreamer(); err != nil {
			panic(err)
		}

		if _, err = oPlayer.Write(frameBuf); err != nil {
			panic(err)
		}
	}

	frameLen := time.Second / time.Duration(streamer.FrameRate)
	handleFrame := func(v FFTResult) {
		if v.Err != nil {
			panic(err)
		}

		for window.Paused {
			time.Sleep(frameLen)
		}

		if window.Skip >= frameLen {
			// skip some of the audio
			if err = audioStreamer(); err != nil {
				panic(err)
			}

			window.Skip -= frameLen
			return
		}

		playAudio()
		window.Show(v.Data)
	}

	// handle the first value in this way, play one frame of audio early
	preFrames := int(streamer.WindowSize / (time.Second / time.Duration(streamer.FrameRate)) / 2)
	log.Info().Int("preFrames", preFrames).Msg("computed pre-frames, awaiting first generation")

	v := <-to
	for i := 0; i < preFrames; i++ {
		playAudio()
	}

	handleFrame(v)

	for v := range to {
		handleFrame(v)
	}

	log.Info().Msg("finished song")
}

func toHz(i, sampleRate int, windowSize time.Duration) float64 {
	N := (float64(sampleRate/2) * (float64(windowSize) / float64(time.Second))) + 1
	return float64((sampleRate+2)*i) / ((N - 1) * 2)
}

func binForI(i, sampleRate, bins int, fmin, fmax, gamma float64, windowSize time.Duration) int {
	return int(math.Pow(float64(toHz(i, sampleRate, windowSize)-fmin)/float64(fmax-fmin), 1/gamma) * float64(bins))
}

func startAt(binI, bins, sampleRate int, fmax, fmin, gamma float64, windowSize time.Duration) int {
	out := int(((math.Pow(float64(binI)/float64(bins), gamma) * (fmax - fmin)) + fmin) * (float64(windowSize) / float64(time.Second)))
	for binForI(out, sampleRate, bins, fmin, fmax, gamma, windowSize) > binI {
		out--
	}

	for binForI(out, sampleRate, bins, fmin, fmax, gamma, windowSize) < binI {
		out++
	}

	return out
}

func dataRange(binI, bins, sampleRate int, fmax, fmin, gamma float64, windowSize time.Duration) (f, l float64, fN, lN int) {
	fN = startAt(binI, bins, sampleRate, fmax, fmin, gamma, windowSize)
	lN = startAt(binI+1, bins, sampleRate, fmax, fmin, gamma, windowSize) - 1
	f = toHz(fN, sampleRate, windowSize)
	l = toHz(lN, sampleRate, windowSize)
	return
}
