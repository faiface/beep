package effects

import "github.com/faiface/beep"

// Loop takes a StreamSeeker and plays it count times. If count is negative, s is looped infinitely.
func Loop(count int, s beep.StreamSeeker) beep.Streamer {
	return &loop{
		s:       s,
		remains: count,
	}
}

type loop struct {
	s       beep.StreamSeeker
	remains int
}

func (l *loop) Stream(samples [][2]float64) (n int, ok bool) {
	if l.remains == 0 || l.s.Err() != nil {
		return 0, false
	}
	for len(samples) > 0 {
		sn, sok := l.s.Stream(samples)
		if !sok {
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
