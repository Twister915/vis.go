package fft

//#cgo LDFLAGS: -lfftw3 -lfftw3_threads
//#include "fftw3.h"
import "C"

import (
	"fmt"
	"runtime"
	"time"
	"unsafe"

	"sync"

	"github.com/rs/zerolog/log"
)

var planMutex = new(sync.Mutex)

func init() {
	C.fftw_init_threads()
}

func fft_plan(timelimit time.Duration, nthreads, n int,
	howmany int,
	input unsafe.Pointer, istride, idist int,
	output unsafe.Pointer, ostride, odist int,
	flags C.uint) C.fftw_plan {

	ns := []int{n}
	log.Info().
		Int("rank", 1).
		Ints("n", ns).
		Int("howmany", howmany).
		Str("in", fmt.Sprintf("%p", input)).
		Int("istride", istride).
		Int("idist", idist).
		Str("out", fmt.Sprintf("%p", output)).
		Int("ostride", ostride).
		Int("odist", odist).
		Int("cpus", nthreads).
		Str("timelimit", timelimit.String()).
		Msg("planning FFT")

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	planMutex.Lock()
	defer planMutex.Unlock()

	if timelimit <= 0 {
		C.fftw_set_timelimit(C.FFTW_NO_TIMELIMIT)
	} else {
		C.fftw_set_timelimit(C.double(float64(timelimit) / float64(time.Second)))
	}

	C.fftw_plan_with_nthreads(C.int(nthreads))

	plan := C.fftw_plan_many_dft_r2c(
		C.int(1),
		(*C.int)(unsafe.Pointer(&ns[0])),
		C.int(howmany),

		(*C.double)(input),
		nil,
		C.int(istride),
		C.int(idist),

		(*C.fftw_complex)(output),
		nil,
		C.int(ostride),
		C.int(odist),

		flags,
	)

	if plan == nil {
		panic("could not construct plan")
	}

	return plan
}
