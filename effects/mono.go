package effects

import "github.com/faiface/beep"

// Mono converts the wrapped Streamer to a mono buffer
// by downmixing the left and right channels together.
type Mono struct {
	Streamer beep.Streamer
}

// Stream streams the wrapped Streamer while converting the channels into mono.
func (m *Mono) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = m.Streamer.Stream(samples)
	for i := range samples[:n] {
		mix := (samples[i][0] + samples[i][1]) / 2
		samples[i][0], samples[i][1] = mix, mix
	}
	return n, ok
}

// Err propagates the wrapped Streamer's errors.
func (m *Mono) Err() error {
	return m.Streamer.Err()
}
