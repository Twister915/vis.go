package audio

import "time"

type SampleReadDirection int

const (
	ReadSampleByChannel SampleReadDirection = iota
	ReadChannelBySample
)

type Input interface {
	BitDepth() int

	Channels() int

	//time between two samples
	Timebase() time.Duration

	SampleRate() int

	Length() time.Duration

	Frames() int

	Seek(int) error

	Reset() error

	ReadSamples([][]float64) (int, error)

	ReadSamplesDir([][]float64, SampleReadDirection) (int, error)

	ReadSample() ([]float64, error)

	ReadNSamples(n int) ([][]float64, error)

	Has(int) bool

	Close() error
}
