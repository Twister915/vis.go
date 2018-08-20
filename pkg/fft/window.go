package fft

import "math"

type WindowingFunction func(n, N float64) float64

func HammingWindow(n, N float64) float64 {
	const (
		alpha = 0.53836
		beta  = 1 - alpha

		TWOPI = math.Pi * 2
	)

	bt := beta * math.Cos((TWOPI*float64(n))/N-1)
	return alpha - bt
}

func BlackmanWindow(n, N float64) float64 {
	const (
		a0 = 0.42659
		a1 = 0.49656
		a2 = 0.076849

		TWOPI  = float64(math.Pi * 2)
		FOURPI = float64(math.Pi * 4)
	)

	nMinusOne := N - 1
	a1t := a1 * math.Cos((TWOPI*float64(n))/nMinusOne)
	a2t := a2 * math.Cos((FOURPI*float64(n))/nMinusOne)
	return a0 - a1t + a2t
}

func BlackmanNuttallWindow(n, N float64) float64 {
	const (
		a0 = 0.3635819
		a1 = 0.4891775
		a2 = 0.1365995
		a3 = 0.0106411

		TWOPI  = float64(math.Pi * 2)
		FOURPI = float64(math.Pi * 4)
		SIXPI  = float64(math.Pi * 6)
	)

	nMinusOne := N - 1
	a1t := a1 * math.Cos((TWOPI*float64(n))/nMinusOne)
	a2t := a2 * math.Cos((FOURPI*float64(n))/nMinusOne)
	a3t := a3 * math.Cos((SIXPI*float64(n))/nMinusOne)
	return a0 - a1t + a2t - a3t
}

func NoWindow(n, N float64) float64 {
	return 1
}
