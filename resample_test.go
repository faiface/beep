package beep_test

import (
	"reflect"
	"testing"

	"github.com/faiface/beep"
)

func TestResample(t *testing.T) {
	for _, numSamples := range []int{8, 100, 500, 1000, 50000} {
		for _, old := range []beep.SampleRate{100, 2000, 44100, 48000} {
			for _, new := range []beep.SampleRate{100, 2000, 44100, 48000} {
				if numSamples/int(old)*int(new) > 1e6 {
					continue // skip too expensive combinations
				}

				s, data := randomDataStreamer(numSamples)

				want := resampleCorrect(3, old, new, data)

				got := collect(beep.Resample(3, old, new, s))

				if !reflect.DeepEqual(want, got) {
					t.Fatal("Resample not working correctly")
				}
			}
		}
	}
}

func resampleCorrect(quality int, old, new beep.SampleRate, p [][2]float64) [][2]float64 {
	ratio := float64(old) / float64(new)
	pts := make([]point, quality*2)
	var resampled [][2]float64
	for i := 0; ; i++ {
		j := float64(i) * ratio
		if int(j) >= len(p) {
			break
		}
		var sample [2]float64
		for c := range sample {
			for k := range pts {
				l := int(j) + k - len(pts)/2 + 1
				if l >= 0 && l < len(p) {
					pts[k] = point{X: float64(l), Y: p[l][c]}
				} else {
					pts[k] = point{X: float64(l), Y: 0}
				}
			}
			y := lagrange(pts[:], j)
			sample[c] = y
		}
		resampled = append(resampled, sample)
	}
	return resampled
}

func lagrange(pts []point, x float64) (y float64) {
	y = 0.0
	for j := range pts {
		l := 1.0
		for m := range pts {
			if j == m {
				continue
			}
			l *= (x - pts[m].X) / (pts[j].X - pts[m].X)
		}
		y += pts[j].Y * l
	}
	return y
}

type point struct {
	X, Y float64
}
