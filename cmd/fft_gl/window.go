package main

import (
	"sync"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

func NewWindow() *window {
	return &window{}
}

type window struct {
	w *glfw.Window

	l         sync.Mutex
	state     []float64
	highWater []float64
	dydtHW    []float64
	lrMaxBuf  []float64

	colors cols
	start  time.Time

	Paused bool
	Skip time.Duration
}

func (w *window) Init(width, height int, fs bool, title string) (err error) {
	w.colors = []col{{253, 255, 183, 255}, {229, 229, 204, 255}}
	w.colors.toDec()

	if err = glfw.Init(); err != nil {
		return
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.AutoIconify, glfw.False)

	var m *glfw.Monitor
	if fs {
		monitors := glfw.GetMonitors()
		m = monitors[len(monitors)-1]
		mode := m.GetVideoMode()
		width, height = mode.Width, mode.Height
	}

	if w.w, err = glfw.CreateWindow(width, height, title, m, nil); err != nil {
		return
	}

	w.w.MakeContextCurrent()
	w.w.SetKeyCallback(w.keyCallback)
	w.w.SetInputMode(glfw.CursorMode, glfw.CursorHidden)

	if err = gl.Init(); err != nil {
		return
	}

	w.initGL()
	return
}

func (w *window) UpdateLoop(rate int) {
	var lastUpdate time.Time
	sleepFor := time.Second / time.Duration(rate)
	for {
		if !lastUpdate.IsZero() {
			time.Sleep(sleepFor - time.Since(lastUpdate))
			w.update(sleepFor)
		}

		lastUpdate = time.Now()
	}
}

func (w *window) DrawLoop() {
	for !w.w.ShouldClose() {
		if w.start.IsZero() {
			w.start = time.Now()
		}

		w.draw()
		w.w.SwapBuffers()
		glfw.PollEvents()
	}
}

func (w *window) Show(frame [][]float64) {
	w.l.Lock()
	defer w.l.Unlock()

	w.showViz(frame)
}

func (w *window) update(delta time.Duration) {
	w.l.Lock()
	defer w.l.Unlock()

	w.updateViz(delta)
}

func (w *window) initGL() {
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(0, 0, 0, 0)

	width, height := w.w.GetSize()
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()

	gl.Ortho(0, float64(width), float64(height), 0, -10, 10)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
}

func (w *window) draw() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	width, height := w.w.GetSize()
	if w.state != nil {
		w.l.Lock()
		defer w.l.Unlock()

		w.drawViz(0, 0, float32(width), float32(height))
	}
}

func (w *window) Close() {
	w.w.SetShouldClose(true)
}

func (w *window) keyCallback(win *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		switch key {
		case glfw.KeyEscape:
			w.w.SetShouldClose(true)
		case glfw.KeySpace:
			w.Paused = !w.Paused
		case glfw.KeyRight:
			w.Skip += time.Second * 5
		}
	}
}