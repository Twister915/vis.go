package main

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestPercentile(t *testing.T) {
	spew.Dump(percentile(29, 6, 0.9))
}
