package wav

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Twister915/vis.go/pkg/audio"
	"github.com/Twister915/vis.go/pkg/util"
	"golang.org/x/exp/mmap"
)

var errShortRead = errors.New("need more bytes (n < min bytes)")

type wavInput struct {
	f      io.ReadSeeker
	header wavHeader

	mutex    *sync.Mutex
	frame    int
	ordering binary.ByteOrder
	closed   bool

	buf []byte

	dataStart int
}

type wavHeader struct {
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	DataSize      uint32
}

func OpenWavMMap(file string) (audio.Input, error) {
	o, err := mmap.Open(file)
	if err != nil {
		return nil, err
	}

	return ReadWav(&util.MMapSeeker{M: o})
}

func OpenWav(file string) (input audio.Input, err error) {
	f, err := os.Open(file)
	if err != nil {
		return
	}

	return ReadWav(f)
}

func OpenWavPreLoad(file string) (input audio.Input, err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	return ReadWav(bytes.NewReader(data))
}

func ReadWav(source io.ReadSeeker) (input audio.Input, err error) {
	wav := new(wavInput)
	wav.mutex = new(sync.Mutex)
	wav.f = source

	if err = wav.readHeader(); err != nil {
		return
	}

	input = wav
	return
}

func (w *wavInput) readHeader() (err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.frame = 0
	// skip back to front of file
	if _, err = w.f.Seek(0, io.SeekStart); err != nil {
		return
	}

	var n int

	// read...
	//
	//  * [4] ChunkID [read, just for validation]
	//  * [4] ChunkSize [skipped]
	//  * [4] Format [checked]
	//
	{
		var chunkID [4]byte
		if n, err = w.f.Read(chunkID[:]); err != nil {
			return
		} else if n != len(chunkID) {
			err = errShortRead
			return
		}

		switch string(chunkID[:]) {
		case "RIFX":
			w.ordering = binary.BigEndian
		case "RIFF":
			w.ordering = binary.LittleEndian
		default:
			err = fmt.Errorf("invalid chunk ID '%s'", string(chunkID[:]))
			return
		}

		if _, err = w.f.Seek(8, io.SeekCurrent); err != nil {
			return
		}
	}

	// read...
	//
	//  * [4] SubChunkID    [read]
	//  * [4] SubChunkSize  [read]
	//  * [2] AudioFormat   [read]
	//  * [2] NumChannels   [read]
	//  * [4] SampleRate    [read]
	//  * [4] ByteRate      [read]
	//  * [2] BlockAlign    [read]
	//  * [2] BitsPerSample [read]
	//  * [?] Anything Else [skipped]
	//
	{
		var subChunkID [4]byte
		if n, err = w.f.Read(subChunkID[:]); err != nil {
			return
		} else if n != len(subChunkID) {
			err = errShortRead
			return
		}

		if string(subChunkID[:]) != "fmt " {
			err = fmt.Errorf("invalid sub-chunk ID '%s'", string(subChunkID[:]))
			return
		}

		var subChunkSize uint32
		if err = binary.Read(w.f, w.ordering, &subChunkSize); err != nil {
			return
		}

		fieldsToRead := []interface{}{
			&w.header.AudioFormat,   //2
			&w.header.NumChannels,   //2
			&w.header.SampleRate,    //4
			&w.header.ByteRate,      //4
			&w.header.BlockAlign,    //2
			&w.header.BitsPerSample, //2
		}

		for _, f := range fieldsToRead {
			if err = binary.Read(w.f, w.ordering, f); err != nil {
				return
			}
		}

		// if not PCM
		if w.header.AudioFormat != 1 {
			if _, err = w.f.Seek(int64(subChunkSize)-16, io.SeekCurrent); err != nil {
				return
			}
		}
	}

	// read...
	//  * [4] SubChunkID   [read]
	//  * [4] SubChunkSize [read]
	//
	for {
		var chunkID [4]byte
		if n, err = w.f.Read(chunkID[:]); err != nil {
			return
		} else if n != len(chunkID) {
			err = errShortRead
			return
		}

		if strings.ToLower(string(chunkID[:])) != "data" {
			//skip this chunk
			var size uint32
			if err = binary.Read(w.f, w.ordering, &size); err != nil {
				if err == io.EOF {
					err = errors.New("no data chunk found")
				}

				return
			}

			if _, err = w.f.Seek(int64(size), io.SeekCurrent); err != nil {
				return
			}
		} else {
			err = binary.Read(w.f, w.ordering, &w.header.DataSize)
			return
		}
	}

	// file should now be pointing at the start of the data

	return
}

func (w *wavInput) BitDepth() int {
	return int(w.header.BitsPerSample)
}

func (w *wavInput) Channels() int {
	return int(w.header.NumChannels)
}

func (w *wavInput) Timebase() time.Duration {
	return time.Second / time.Duration(w.header.SampleRate)
}

func (w *wavInput) SampleRate() int {
	return int(w.header.SampleRate)
}

func (w *wavInput) Frames() int {
	return int(w.header.DataSize) / int(w.header.NumChannels*(w.header.BitsPerSample/8))
}

func (w *wavInput) Length() time.Duration {
	return time.Duration(w.Frames()) * w.Timebase()
}

func (w *wavInput) ReadSamples(to [][]float64) (n int, err error) {
	return w.ReadSamplesDir(to, audio.ReadSampleByChannel)
}

func (w *wavInput) ReadSamplesDir(to [][]float64, dir audio.SampleReadDirection) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	switch dir {
	case audio.ReadSampleByChannel:
		if len(to[0]) != int(w.header.NumChannels) {
			err = errors.New("must pass [][]float64 pre-allocated with correct number of channels")
			return
		}
	case audio.ReadChannelBySample:
		if len(to) != int(w.header.NumChannels) {
			err = errors.New("must pass [][]float64 pre-allocated with correct number of channels")
			return
		}
	default:
		panic("invalid dir")
	}

	if w.closed {
		panic("read from closed file")
	}

	if dir == audio.ReadSampleByChannel {
		n = len(to)
	} else {
		n = len(to[0])
	}

	{
		frames := w.Frames()
		end := w.frame + n
		if end > frames {
			n = frames - w.frame
		}
	}

	if n <= 0 {
		err = io.EOF
		return
	}

	size := int(w.header.BlockAlign) * n
	var buf []byte
	if len(w.buf) == size {
		buf = w.buf
	} else if len(w.buf) > size {
		buf = w.buf[:size]
	} else {
		w.buf = make([]byte, size)
		buf = w.buf
	}

	var bn int
	if bn, err = io.ReadFull(w.f, buf); err != nil {
		return
	} else if bn < len(buf) {
		err = errShortRead
		return
	}

	dataI := 0
	for i := 0; i < n; i++ {
		for z := 0; z < int(w.header.NumChannels); z++ {
			var v float64
			switch int(w.header.BitsPerSample) {
			case 8:
				v = float64(int8(buf[dataI])) / (float64(1<<7) - 1)
				dataI++
			case 16:
				const bytes16 = 2
				twoBits := buf[dataI : dataI+bytes16]
				vUint := w.ordering.Uint16(twoBits)
				vInt := int16(vUint)
				v = float64(vInt) / (float64(1<<15) - 1)
				dataI += bytes16
			default:
				err = fmt.Errorf("no support for this bits per sample (%d)", w.header.BitsPerSample)
				return
			}

			switch dir {
			case audio.ReadSampleByChannel:
				to[i][z] = v
			case audio.ReadChannelBySample:
				to[z][i] = v
			}
		}

		w.frame++
	}

	return
}

func (w *wavInput) ReadNSamples(n int) (out [][]float64, err error) {
	out = util.Create2DFloats(n, int(w.header.NumChannels))
	read, err := w.ReadSamples(out)
	if err != nil {
		return
	}

	if read < n {
		err = io.EOF
	}

	return
}

func (w *wavInput) ReadSample() (out []float64, err error) {
	samples, err := w.ReadNSamples(1)
	if err != nil {
		return
	}

	out = samples[0]
	return
}

func (w *wavInput) Close() (err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	defer func() {
		if err == nil {
			w.closed = true
		}
	}()

	if closer, ok := w.f.(io.Closer); ok {
		err = closer.Close()
	}

	return
}

func (w *wavInput) Has(n int) bool {
	return (w.Frames() - w.frame) >= n
}

func (w *wavInput) Seek(n int) (err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	futureFrame := w.frame + n
	if futureFrame < 0 || futureFrame >= w.Frames() {
		err = io.EOF
		return
	}

	bytesSeek := int64(n * int(w.header.NumChannels) * int(w.header.BitsPerSample/8))
	if _, err = w.f.Seek(bytesSeek, io.SeekCurrent); err != nil {
		return
	}

	w.frame += n
	return
}

func (w *wavInput) Reset() (err error) {
	return w.readHeader()
}
