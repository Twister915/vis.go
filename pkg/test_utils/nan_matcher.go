package test_utils

import (
	"math"

	"fmt"

	"github.com/onsi/gomega"
)

func BeNaN() gomega.OmegaMatcher {
	return nanMatcher{}
}

type nanMatcher struct{}

func (n nanMatcher) Match(actual interface{}) (success bool, err error) {
	switch v := actual.(type) {
	case float32:
		success = math.IsNaN(float64(v))
	case float64:
		success = math.IsNaN(v)
	case complex64:
		success = math.IsNaN(float64(real(v))) || math.IsNaN(float64(imag(v)))
	case complex128:
		success = math.IsNaN(real(v)) || math.IsNaN(imag(v))
	default:
		err = fmt.Errorf("%v is not a float", actual)
	}

	return
}

func (n nanMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("%v is not NaN", actual)
}

func (n nanMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("%v is NaN", actual)
}
