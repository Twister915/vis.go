package util

import "reflect"

func IterateElements(in interface{}, f func([]int, interface{})) {
	itr_elem(reflect.ValueOf(in), f, []int{})
}

func itr_elem(in reflect.Value, f func([]int, interface{}), is []int) {
	if in.Kind() != reflect.Slice {
		f(is, in.Interface())
	} else {
		for i := 0; i < in.Len(); i++ {
			itr_elem(in.Index(i), f, append(is, i))
		}
	}
}

func UpdateElements(in interface{}, f func([]int, interface{}) interface{}) {
	update_elem(reflect.ValueOf(in), f, []int{})
}

func update_elem(in reflect.Value, f func([]int, interface{}) interface{}, is []int) {
	iT := in.Type()
	isSlice := iT.Kind() == reflect.Slice
	isPointerToSlice := iT.Kind() == reflect.Ptr && iT.Elem().Kind() == reflect.Slice

	if !isSlice && !isPointerToSlice {
		in.Elem().Set(reflect.ValueOf(f(is, in.Elem().Interface())))
	} else {
		if isPointerToSlice {
			in = in.Elem()
		} else if !isSlice {
			panic("bad argument passed")
		}

		for i := 0; i < in.Len(); i++ {
			update_elem(in.Index(i).Addr(), f, append(is, i))
		}
	}
}
