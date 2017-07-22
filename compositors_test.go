package beep_test

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/faiface/beep"
)

// randomDataStreamer generates random samples of duration d and returns a Streamer which streams
// them and the data itself.
func randomDataStreamer(d time.Duration) (s beep.Streamer, data [][2]float64) {
	numSamples := int(math.Ceil(d.Seconds() * float64(beep.SampleRate)))
	data = make([][2]float64, numSamples)
	for i := range data {
		data[i][0] = rand.Float64()*2 - 1
		data[i][1] = rand.Float64()*2 - 1
	}
	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		if len(data) == 0 {
			return 0, false
		}
		n = copy(samples, data)
		data = data[n:]
		return n, true
	}), data
}

// collect drains Streamer s and returns all of the samples it streamed.
func collect(s beep.Streamer) [][2]float64 {
	var (
		result [][2]float64
		buf    [512][2]float64
	)
	for {
		n, ok := s.Stream(buf[:])
		if !ok {
			return result
		}
		result = append(result, buf[:n]...)
	}
}

func TestTake(t *testing.T) {
	for i := 0; i < 7; i++ {
		total := time.Nanosecond * time.Duration(1e8+rand.Intn(1e9))
		s, data := randomDataStreamer(total)
		d := time.Nanosecond * time.Duration(rand.Int63n(total.Nanoseconds()))
		numSamples := int(math.Ceil(d.Seconds() * float64(beep.SampleRate)))

		want := data[:numSamples]
		got := collect(beep.Take(d, s))

		if !reflect.DeepEqual(want, got) {
			t.Error("Take not working correctly")
		}
	}
}

func TestSeq(t *testing.T) {
	var (
		n    = 7
		s    = make([]beep.Streamer, n)
		data = make([][][2]float64, n)
	)
	for i := range s {
		s[i], data[i] = randomDataStreamer(time.Nanosecond * time.Duration(1e8+rand.Intn(1e9)))
	}

	var want [][2]float64
	for _, d := range data {
		want = append(want, d...)
	}

	got := collect(beep.Seq(s...))

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Seq not working properly")
	}
}

func TestMix(t *testing.T) {
	var (
		n    = 7
		s    = make([]beep.Streamer, n)
		data = make([][][2]float64, n)
	)
	for i := range s {
		s[i], data[i] = randomDataStreamer(time.Nanosecond * time.Duration(1e8+rand.Intn(1e9)))
	}

	maxLen := 0
	for _, d := range data {
		if len(d) > maxLen {
			maxLen = len(d)
		}
	}

	want := make([][2]float64, maxLen)
	for _, d := range data {
		for i := range d {
			want[i][0] += d[i][0]
			want[i][1] += d[i][1]
		}
	}

	got := collect(beep.Mix(s...))

	if !reflect.DeepEqual(want, got) {
		t.Error("Mix not working correctly")
	}
}
