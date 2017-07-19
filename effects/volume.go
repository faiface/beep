package effects

import (
	"math"

	"github.com/faiface/beep"
)

type Volume struct {
	Streamer beep.Streamer
	Base     float64
	Volume   float64
	Silent   bool
}

func (v *Volume) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = v.Streamer.Stream(samples)
	gain := 0.0
	if !v.Silent {
		gain = math.Pow(v.Base, v.Volume)
	}
	for i := range samples[:n] {
		samples[i][0] *= gain
		samples[i][1] *= gain
	}
	return n, ok
}

func (v *Volume) Err() error {
	return v.Streamer.Err()
}
