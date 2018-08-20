package main

import (
	"github.com/Twister915/vis.go/pkg/audio"
	"github.com/Twister915/vis.go/pkg/util"
)

type player struct {
	input audio.Input
}

func (p *player) Stream(bitDepth int, to []byte) func() (err error) {
	samples := len(to) / ((bitDepth / 8) * p.input.Channels())
	buf := util.Create2DFloats(samples, p.input.Channels())
	return func() (err error) {
		if _, err = p.input.ReadSamples(buf); err != nil {
			return
		}

		bitsPerChannel := bitDepth / 8
		bytesPerSample := bitsPerChannel * p.input.Channels()
		for i := range buf {
			for c, val := range buf[i] {
				switch bitDepth {
				case 8:
					to[i*bytesPerSample+c*bitsPerChannel] = byte(int16(val * (1<<7 - 1)))
				case 16:
					valInt16 := int16(val * (1<<15 - 1))
					to[i*bytesPerSample+c*bitsPerChannel+0] = byte(valInt16)
					to[i*bytesPerSample+c*bitsPerChannel+1] = byte(valInt16 >> 8)
				}
			}
		}

		return
	}
}
