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
