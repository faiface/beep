package beep

// Silence returns a Streamer which streams num samples of silence. If num is negative, silence is
// streamed forever.
func Silence(num int) Streamer {
	return StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		if num == 0 {
			return 0, false
		}
		if num < len(samples) {
			samples = samples[:num]
		}
		for i := range samples {
			samples[i] = [2]float64{}
		}
		num -= len(samples)
		return len(samples), true
	})
}

// Callback returns a Streamer, which does not stream any samples, but instead calls f the first
// time its Stream method is called.
func Callback(f func()) Streamer {
	return StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		if f != nil {
			f()
			f = nil
		}
		return 0, false
	})
}
