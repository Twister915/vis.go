package fft

//#include "fftw3.h"
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/Twister915/vis.go/pkg/util"
)

type FFTWDoubles3D [][][]float64

func Alloc3DDoubles(x, y, z int) FFTWDoubles3D {
	backing := AllocFFTWDoubles(x * y * z)
	return FFTWDoubles3D(util.RearrangeNDTs(backing, x, y, z).([][][]float64))
}

func (f FFTWDoubles3D) Free() {
	C.fftw_free(unsafe.Pointer(&f[0][0][0]))
}

type FFTWComplexes3D [][][]complex128

func Alloc3DComplexes(x, y, z int) FFTWComplexes3D {
	backing := AllocFFTWComplexes(x * y * z)
	return FFTWComplexes3D(util.RearrangeNDTs(backing, x, y, z).([][][]complex128))
}

func (f FFTWComplexes3D) Free() {
	C.fftw_free(unsafe.Pointer(&f[0][0][0]))
}

type FFTWDoubles2D [][]float64

func Alloc2DDoubles(x, y int) FFTWDoubles2D {
	backing := AllocFFTWDoubles(x * y)
	return FFTWDoubles2D(util.Reshape2DFloats(x, y, backing))
}

func (f FFTWDoubles2D) Free() {
	C.fftw_free(unsafe.Pointer(&f[0][0]))
}

type FFTWComplexes2D [][]complex128

func Alloc2DComplexes(x, y int) FFTWComplexes2D {
	backing := AllocFFTWComplexes(x * y)
	return FFTWComplexes2D(util.Reshape2DComplex(x, y, backing))
}

func (f FFTWComplexes2D) Free() {
	C.fftw_free(unsafe.Pointer(&f[0][0]))
}

type FFTWDoubles []float64

func AllocFFTWDoubles(n int) FFTWDoubles {
	return FFTWDoubles(*((*[]float64)(fftw_malloc_slice(uintptr(n)*unsafe.Sizeof(float64(0)), n))))
}

func (f FFTWDoubles) Free() {
	C.fftw_free(unsafe.Pointer(&f[0]))
}

type FFTWComplexes []complex128

func AllocFFTWComplexes(n int) FFTWComplexes {
	return FFTWComplexes(*((*[]complex128)(fftw_malloc_slice(uintptr(n)*unsafe.Sizeof(complex128(0)), n))))
}

func (f FFTWComplexes) Free() {
	C.fftw_free(unsafe.Pointer(&f[0]))
}

func fftw_malloc(size uintptr) unsafe.Pointer {
	fmt.Printf("[alloc] fftw_malloc(%d)\n", int(size))
	return C.fftw_malloc(C.size_t(size))
}

func fftw_malloc_slice(size uintptr, len int) unsafe.Pointer {
	return unsafe.Pointer(&reflect.SliceHeader{
		Len:  len,
		Cap:  len,
		Data: uintptr(fftw_malloc(size)),
	})
}

func CheckAlignment(as ...interface{}) bool {
	a := -1
	for _, v := range as {
		aV := int(C.fftw_alignment_of((*C.double)(resolveHeadPtr(reflect.ValueOf(v)))))
		if a == -1 {
			a = aV
		}

		if a != aV {
			return false
		}
	}

	return true
}

func resolveHeadPtr(in reflect.Value) unsafe.Pointer {
	if in.Kind() == reflect.Ptr && in.Type().Elem().Kind() == reflect.Slice {
		return resolveHeadPtr(in.Elem())
	}

	if in.Kind() != reflect.Slice {
		panic("must pass slice")
	}

	if in.Len() == 0 {
		panic("empty slice")
	}

	if in.Type().Elem().Kind() == reflect.Slice {
		return resolveHeadPtr(in.Index(0))
	}

	return unsafe.Pointer(in.Index(0).Addr().Pointer())
}
