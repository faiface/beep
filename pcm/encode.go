package pcm

import (
	"bufio"
	"io"

	"github.com/faiface/beep"
)

// Encode writes all audio streamed from s to w in raw PCM format.
func Encode(w io.Writer, s beep.Streamer, format beep.Format) error {
	var (
		bw      = bufio.NewWriter(w)
		samples = make([][2]float64, 512)
		buffer  = make([]byte, len(samples)*format.Width())
	)
	for {
		n, ok := s.Stream(samples)
		if !ok {
			return bw.Flush()
		}
		var offset int
		for _, sample := range samples[:n] {
			offset += format.EncodeSigned(buffer[offset:], sample)
		}
		if _, err := bw.Write(buffer[:offset]); err != nil {
			return err
		}
	}
}

// NewReader returns an io.Reader of the audio stream formatted as raw PCM.
func NewReader(s beep.Streamer, format beep.Format, bufferSize int) io.Reader {
	return &reader{
		s:       s,
		f:       format,
		samples: make([][2]float64, bufferSize),
	}
}

type reader struct {
	s       beep.Streamer
	f       beep.Format
	samples [][2]float64
	len     int
	pos     int
}

// Read implements io.Reader
func (r *reader) Read(data []byte) (n int, err error) {
	// get more samples if the are none left
	if r.len == r.pos {
		var ok bool
		r.pos = 0
		r.len, ok = r.s.Stream(r.samples)
		if !ok {
			if err := r.s.Err(); err != nil {
				return 0, err
			}
			return 0, io.EOF
		}
	}
	// figure out how many samples we're decoding
	nsample := r.len - r.pos
	if max := len(data) / r.f.Width(); max < nsample {
		nsample = max
	}
	// do the decoding
	for i := 0; i < nsample; i++ {
		n += r.f.EncodeSigned(data[n:], r.samples[r.pos+i])
	}
	r.pos += nsample
	return n, nil
}
