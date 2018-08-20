package util

import "reflect"

func CreateNDTs(tE interface{}, vs ...int) interface{} {
	if len(vs) == 0 {
		return nil
	}

	N := MultiplyAll(vs...)
	return RearrangeNDTs(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(tE)), N, N).Interface(), vs...)
}

func RearrangeNDTs(backingRaw interface{}, vs ...int) interface{} {
	if len(vs) == 0 {
		return nil
	}

	backing := reflect.ValueOf(backingRaw)

	if backing.Len() != MultiplyAll(vs...) {
		panic("bac backing array passed")
	}

	t := backing.Type().Elem()

	n := 0

	var d func(reflect.Value, ...int)
	d = func(to reflect.Value, ss ...int) {
		if len(ss) == 1 {
			size := ss[0]

			for i := 0; i < to.Len(); i++ {
				to.Index(i).Set(backing.Slice(n, n+size))
				n += size
			}

			return
		}

		myType := to.Type().Elem()
		for i := 0; i < to.Len(); i++ {
			to.Index(i).Set(reflect.MakeSlice(myType, ss[0], ss[0]))
		}

		for i := 0; i < to.Len(); i++ {
			d(to.Index(i), ss[1:]...)
		}
	}

	// compute the highest slice type
	for i := 0; i < len(vs); i++ {
		t = reflect.SliceOf(t)
	}

	// create the highest slice
	out := reflect.MakeSlice(t, vs[0], vs[0])
	// fill slice with smaller slices of lower types
	d(out, vs[1:]...)

	return out.Interface()
}

func CreateNDFloat64(vs ...int) interface{} {
	return CreateNDTs(float64(0), vs...)
}

func CreateNDComplex128(vs ...int) interface{} {
	return CreateNDTs(complex128(0), vs...)
}
