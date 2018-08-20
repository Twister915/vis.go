package test_utils

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/onsi/gomega"
)

func BeContiguous() gomega.OmegaMatcher {
	return contiguousMatcher{}
}

type contiguousMatcher struct{}

func (cm contiguousMatcher) Match(arr interface{}) (success bool, err error) {
	a := reflect.ValueOf(arr)
	if a.Kind() != reflect.Slice {
		err = errors.New("must be a slice to do contiguous testing on")
	}

	success = cm.doTest(0, 0, a) == nil
	return
}

func (cm contiguousMatcher) doTest(p, d int, a reflect.Value) []int {
	isHD := a.Type().Elem().Kind() == reflect.Slice
	for i := 0; i < a.Len(); i++ {
		if !isHD {
			ptr := int(a.Index(i).Addr().Pointer())
			isFirst := p == 0
			if isFirst {
				p = ptr
			} else if d == 0 {
				d = ptr - p
			}

			if !isFirst && ptr != p+d*i {
				return []int{i}
			}
		} else if fI := cm.doTest(p, d, a.Index(i)); fI != nil {
			return append([]int{i}, fI...)
		}
	}

	if !isHD {
		p += d * a.Len()
	}

	return nil
}

func (cm contiguousMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("slice is not contigious at %v", cm.doTest(0, 0, reflect.ValueOf(actual)))
}

func (cm contiguousMatcher) NegatedFailureMessage(actual interface{}) string {
	return "slice is contiguous!"
}
