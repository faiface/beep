// tones generator

package generators

import (
	"math"
	"errors"
	. "github.com/faiface/beep"
)

// simple sinusoid tone generator
type toneStreamer struct {
	stat  float64
	delta float64
}

// create streamer which will produce infinite sinusoid tone with the given frequency
// use other wrappers of this package to change amplitude or add time limit
// sampleRate must be at least two times grater then frequency, otherwise this function will return an error
func SinTone(sr SampleRate, freq int) (Streamer, error) {
	if int(sr)/freq < 2 {
		return nil, errors.New("faiface beep tone generator: samplerate must be at least 2 times grater then frequency")
	}
	r := new(toneStreamer)
	r.stat = 0.0
	srf := float64(sr)
	ff := float64(freq)
	steps := srf / ff
	r.delta = 1.0 / steps
	return r, nil
}

func (c *toneStreamer) nextSample() float64 {
	r := math.Sin(c.stat * 2.0 * math.Pi)
	_, c.stat = math.Modf(c.stat + c.delta)
	return r
}

func (c *toneStreamer) Stream(buf [][2]float64) (int, bool) {
	for i := 0; i < len(buf); i++ {
		s := c.nextSample()
		buf[i] = [2]float64{s, s}
	}
	return len(buf), true
}
func (_ *toneStreamer) Err() error {
	return nil
}
