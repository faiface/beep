package beep

// Take returns a Streamer which streams at most n samples from s.
//
// The returned Streamer propagates s's errors throught Err.
func Take(n int, s Streamer) Streamer {
	return &take{
		s:          s,
		currSample: 0,
		numSamples: n,
	}
}

type take struct {
	s          Streamer
	currSample int
	numSamples int
}

func (t *take) Stream(samples [][2]float64) (n int, ok bool) {
	if t.currSample >= t.numSamples {
		return 0, false
	}
	toStream := t.numSamples - t.currSample
	if len(samples) < toStream {
		toStream = len(samples)
	}
	n, ok = t.s.Stream(samples[:toStream])
	t.currSample += n
	return n, ok
}

func (t *take) Err() error {
	return t.s.Err()
}

// Loop takes a StreamSeeker and plays it count times. If count is negative, s is looped infinitely.
//
// The returned Streamer propagates s's errors.
func Loop(count int, s StreamSeeker) Streamer {
	return &loop{
		s:       s,
		remains: count,
	}
}

type loop struct {
	s       StreamSeeker
	remains int
}

func (l *loop) Stream(samples [][2]float64) (n int, ok bool) {
	if l.remains == 0 || l.s.Err() != nil {
		return 0, false
	}
	for len(samples) > 0 {
		sn, sok := l.s.Stream(samples)
		if !sok {
			if l.remains == 0 {
				break
			}
			err := l.s.Seek(0)
			if err != nil {
				return n, true
			}
			if l.remains > 0 {
				l.remains--
			}
			continue
		}
		samples = samples[sn:]
		n += sn
	}
	return n, true
}

func (l *loop) Err() error {
	return l.s.Err()
}

// Seq takes zero or more Streamers and returns a Streamer which streams them one by one without pauses.
//
// Seq does not propagate errors from the Streamers.
func Seq(s ...Streamer) Streamer {
	i := 0
	return StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		for i < len(s) && len(samples) > 0 {
			sn, sok := s[i].Stream(samples)
			samples = samples[sn:]
			n, ok = n+sn, ok || sok
			if !sok {
				i++
			}
		}
		return n, ok
	})
}

// Mix takes zero or more Streamers and returns a Streamer which streames them mixed together.
//
// Mix does not propagate errors from the Streamers.
func Mix(s ...Streamer) Streamer {
	return StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		var tmp [512][2]float64

		for len(samples) > 0 {
			toStream := len(tmp)
			if toStream > len(samples) {
				toStream = len(samples)
			}

			// clear the samples
			for i := range samples[:toStream] {
				samples[i] = [2]float64{}
			}

			snMax := 0 // max number of streamed samples in this iteration
			for _, st := range s {
				// mix the stream
				sn, sok := st.Stream(tmp[:toStream])
				if sn > snMax {
					snMax = sn
				}
				ok = ok || sok

				for i := range tmp[:sn] {
					samples[i][0] += tmp[i][0]
					samples[i][1] += tmp[i][1]
				}
			}

			n += snMax
			if snMax < len(tmp) {
				break
			}
			samples = samples[snMax:]
		}

		return n, ok
	})
}
