package effects

import (
	"math"

	"github.com/faiface/beep"
)

// https://octovoid.com/2017/11/04/coding-a-parametric-equalizer-for-audio-applications/
type (
	section struct {
		a, b         []float64
		xPast, yPast [][2]float64
	}

	EqualizerSection struct {
		F0, Bf, GB, G0, G float64
	}

	equalizer struct {
		streamer beep.Streamer
		sections []section
	}
)

func (s *section) apply(x [][2]float64) [][2]float64 {
	ord := len(s.a) - 1
	np := len(x) - 1

	if np < ord {
		x = append(x, make([][2]float64, ord-np)...)
		np = ord
	}

	y := make([][2]float64, len(x))

	if len(s.xPast) == 0 {
		s.xPast = make([][2]float64, len(x))
	}

	if len(s.yPast) == 0 {
		s.yPast = make([][2]float64, len(x))
	}

	for i := 0; i < len(x); i++ {
		for j := 0; j < ord+1; j++ {
			if i-j < 0 {
				y[i][0] = y[i][0] + s.b[j]*s.xPast[len(s.xPast)-j][0]
				y[i][1] = y[i][1] + s.b[j]*s.xPast[len(s.xPast)-j][1]
			} else {
				y[i][0] = y[i][0] + s.b[j]*x[i-j][0]
				y[i][1] = y[i][1] + s.b[j]*x[i-j][1]
			}
		}

		for j := 0; j < ord; j++ {
			if i-j-1 < 0 {
				y[i][0] = y[i][0] - s.a[j+1]*s.yPast[len(s.yPast)-j-1][0]
				y[i][1] = y[i][1] - s.a[j+1]*s.yPast[len(s.yPast)-j-1][1]
			} else {
				y[i][0] = y[i][0] - s.a[j+1]*y[i-j-1][0]
				y[i][1] = y[i][1] - s.a[j+1]*y[i-j-1][1]
			}

			y[i][0] = y[i][0] / s.a[0]
			y[i][1] = y[i][1] / s.a[0]
		}
	}

	s.xPast = x
	s.yPast = y
	return y
}

func NewEqualizer(s beep.Streamer, fs float64, sections []EqualizerSection) beep.Streamer {
	out := &equalizer{
		streamer: s,
	}

	for _, s := range sections {
		beta := math.Tan(s.Bf/2.0*math.Pi/(fs/2.0)) *
			math.Sqrt(math.Abs(math.Pow(math.Pow(10, s.GB/20.0), 2.0)-
				math.Pow(math.Pow(10.0, s.G0/20.0), 2.0))) /
			math.Sqrt(math.Abs(math.Pow(math.Pow(10.0, s.G/20.0), 2.0)-
				math.Pow(math.Pow(10.0, s.GB/20.0), 2.0)))

		b := []float64{
			(math.Pow(10.0, s.G0/20.0) + math.Pow(10.0, s.G/20.0)*beta) / (1 + beta),
			(-2 * math.Pow(10.0, s.G0/20.0) * math.Cos(s.F0*math.Pi/(fs/2.0))) / (1 + beta),
			(math.Pow(10.0, s.G0/20) - math.Pow(10.0, s.G/20.0)*beta) / (1 + beta),
		}

		a := []float64{
			1.0,
			-2 * math.Cos(s.F0*math.Pi/(fs/2.0)) / (1 + beta),
			(1 - beta) / (1 + beta),
		}
		out.sections = append(out.sections, section{a: a, b: b})
	}
	return out
}

// Stream streams the wrapped Streamer modified by Equalizer.
func (e *equalizer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = e.streamer.Stream(samples)
	for _, s := range e.sections {
		copy(samples, s.apply(samples))
	}
	return n, ok
}

// Err propagates the wrapped Streamer's errors.
func (e *equalizer) Err() error {
	return e.streamer.Err()
}
