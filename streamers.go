package beep

// Silence returns a Streamer which streams n samples of silence. If n is negative, silence is
// streamed forever.
func Silence(n int) Streamer {
	return StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		if n == 0 {
			return n, false
		}
		for i := range samples {
			if n == 0 {
				break
			}
			samples[i] = [2]float64{}
			if n > 0 {
				n--
			}
		}
		return n, true
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
