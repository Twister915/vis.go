package main

import (
	"time"

	"github.com/Twister915/vis.go/pkg/bin"
	"github.com/go-gl/gl/v2.1/gl"
)

const (
	paddingTop    = 20
	paddingBeside = 5

	paddingBetween  = 3
	highWaterHeight = 2

	min = 0.01
)

func (w *window) showViz(frame [][]float64) {
	if w.state == nil {
		w.state = make([]float64, len(frame[0]))
		w.highWater = make([]float64, len(frame[0]))
		w.dydtHW = make([]float64, len(frame[0]))
		w.lrMaxBuf = make([]float64, len(frame[0]))
	}

	bin.CombineChannelsAvg(frame[:], w.state)

	copy(w.lrMaxBuf, frame[0])
	for i, v := range frame[1] {
		if v > w.lrMaxBuf[i] {
			w.lrMaxBuf[i] = v
		}
	}

	for i, v := range w.lrMaxBuf {
		if w.highWater[i] < v {
			w.highWater[i] = v
			w.dydtHW[i] = 0
		}
	}
}

func (w *window) updateViz(delta time.Duration) {
	const fallAcc = 0.015

	fallDvDt := fallAcc * (float64(delta) / float64(time.Second))
	for i, v := range w.highWater {
		if v > w.lrMaxBuf[i] {
			w.dydtHW[i] += fallDvDt
			w.highWater[i] -= w.dydtHW[i]
		} else {
			w.dydtHW[i] = 0
		}
	}
}

func (w *window) drawViz(x, y, wi, he float32) {
	gl.PushMatrix()
	gl.Translatef(x+paddingBeside+paddingBetween, y+he, 0)

	barWidth := ((wi - (paddingBeside * 2)) / float32(len(w.state))) - paddingBetween
	barHeight := he - paddingTop

	w.drawHighChannelLines(barWidth, barHeight)
	w.drawWhiteLines(barWidth, barHeight)
	w.drawHighWaters(barWidth, barHeight)
	gl.PopMatrix()
}

func (w *window) drawWhiteLines(barWidth, barHeight float32) {
	w.drawLines(barWidth, barHeight, w.colors[1], w.state)
}

func (w *window) drawHighChannelLines(barWidth, barHeight float32) {
	w.drawLines(barWidth, barHeight, w.colors[0], w.lrMaxBuf)
}

func (w *window) drawLines(barWidth, barHeight float32, c col, ls []float64) {
	gl.PushMatrix()

	for _, v := range ls {
		w.drawLine(barWidth, barHeight, c, w.normLineHeight(v))
	}

	gl.PopMatrix()
}

func (w *window) drawHighWaters(barWidth, barHeight float32) {
	gl.PushMatrix()
	c := col{255, 174, 0, 255}.toDec()

	for _, v := range w.highWater {
		v := w.normLineHeight(v)
		if v > min {
			w.drawHighWater(barWidth, barHeight, c, v)
		}

		gl.Translatef(float32(barWidth+paddingBetween), 0, 0)
	}

	gl.PopMatrix()
}

func (w *window) normLineHeight(v float64) float32 {
	if v < min {
		v = min
	}

	if v > 1 {
		v = 1
	}

	// discrete version of this, to see what the LEDs will look like
	//const levels = 24
	//v = float64(util.Round(v * float64(levels))) / float64(levels)

	return float32(v)
}

func (w *window) drawHighWater(barWidth, barHeight float32, c col, hw float32) {
	hwY := hw * barHeight
	gl.Color4f(c[0], c[1], c[2], c[3])
	gl.Begin(gl.QUADS)
	gl.Vertex2f(0, -hwY)
	gl.Vertex2f(float32(barWidth), -hwY)
	gl.Vertex2f(float32(barWidth), -(hwY + float32(highWaterHeight)))
	gl.Vertex2f(0, -(hwY + float32(highWaterHeight)))
	gl.End()
}

func (w *window) drawLine(barWidth, barHeight float32, c col, b float32) {
	gl.Color4f(c[0], c[1], c[2], c[3])
	gl.Begin(gl.QUADS)
	gl.Vertex2f(0, 0)
	gl.Vertex2f(barWidth, 0)
	h := -barHeight * b
	gl.Vertex2f(barWidth, h)
	gl.Vertex2f(0, h)
	gl.End()
	gl.Translatef(float32(barWidth+paddingBetween), 0, 0)
}
