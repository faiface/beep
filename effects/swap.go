package effects

import "github.com/faiface/beep"

// Swap swaps the left and right channel of the wrapped Streamer.
type Swap struct {
	Streamer beep.Streamer
}

// Stream streams the wrapped Streamer while having its left and right
// channels swapped.
func (s *Swap) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = s.Streamer.Stream(samples)
	for i := range samples[:n] {
		samples[i][0], samples[i][1] = samples[i][1], samples[i][0]
	}
	return n, ok
}

// Err propagates the wrapped Streamer's errors.
func (s *Swap) Err() error {
	return s.Streamer.Err()
}
