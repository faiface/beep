package pcm

import (
	"io"

	"github.com/faiface/beep"
)

// Decode takes a Reader containing audio data in raw PCM format and returns a Streamer,
// which streams that audio.
func Decode(r io.Reader, format beep.Format) beep.Streamer {
	return &stream{
		r:   r,
		f:   format,
		buf: make([]byte, 512*format.Width()),
	}
}

type stream struct {
	r   io.Reader
	f   beep.Format
	buf []byte
	len int
	pos int
	err error
}

func (s *stream) Err() error { return s.err }

func (s *stream) Stream(samples [][2]float64) (n int, ok bool) {
	width := s.f.Width()
	// if there's not enough data for a full sample, get more
	if size := s.len - s.pos; size < width {
		// if there's a partial sample, move it to the beginning of the buffer
		if size != 0 {
			copy(s.buf, s.buf[s.pos:s.len])
		}
		s.len = size
		s.pos = 0
		// refill the buffer
		nbytes, err := s.r.Read(s.buf[s.len:])
		if err != nil {
			if err != io.EOF {
				s.err = err
			}
			return n, false
		}
		s.len += nbytes
	}
	// decode as many samples as we can
	for n < len(samples) && s.len-s.pos >= width {
		samples[n], _ = s.f.DecodeSigned(s.buf[s.pos:])
		n++
		s.pos += width
	}
	return n, true
}
