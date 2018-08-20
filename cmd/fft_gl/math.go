package main

import (
	"math"

	"github.com/Twister915/vis.go/pkg/util"
)

func analyzeDistribution(data interface{}) (mu, sigma float64) {
	sum := 0.0
	count := 0
	util.IterateElements(data, func(_ []int, d interface{}) {
		v := d.(float64)
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return
		}

		sum += v
		count++
	})

	mu = sum / float64(count)
	squaredResiduals := 0.0
	util.IterateElements(data, func(_ []int, d interface{}) {
		v := d.(float64)
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return
		}

		resid := v - mu
		squaredResiduals += resid * resid
	})

	sigma = math.Sqrt(squaredResiduals / float64(count-1))
	return
}

func percentile(mu, sigma, percentile float64) float64 {
	return mu + sigma*(math.Sqrt(2)*util.ErfInv((2*percentile)-1))
}

func normalizeExtrema(data []float64, mu, sigma, pL, pH float64) {
	normalizeAbout(data, percentile(mu, sigma, pL), percentile(mu, sigma, 1-pH))
}

func normalizeAbout(data interface{}, min, max float64) {
	mm := max - min
	util.UpdateElements(data, func(_ []int, d interface{}) interface{} {
		return (d.(float64) - min) / mm
	})
}

func savitskyGolaySmooth(data []float64, passes, smoothingPoints int) {
	lastArray := data
	side := smoothingPoints / 2
	cn := 1.0 / float64(smoothingPoints+1)
	for pass := 0; pass < passes; pass++ {
		newArray := make([]float64, len(data))
		for i := 0; i < side; i++ {
			newArray[i] = lastArray[i]
			newArray[len(lastArray)-i-1] = lastArray[len(lastArray)-i-1]
		}

		for i := side; i < len(lastArray)-side; i++ {
			sum := 0.0
			for n := -side; n <= side; n++ {
				sum += cn*lastArray[i+n] + float64(n)
			}

			newArray[i] = sum
		}

		lastArray = newArray
	}

	copy(data, lastArray)
}

func IntAbs(i int) int {
	if i < 0 {
		return -i
	}

	return i
}
