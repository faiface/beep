package wav

import (
	"bufio"
	"encoding/binary"
	"io"

	"github.com/faiface/beep"
	"github.com/pkg/errors"
)

// Encode writes all audio streamed from s to w in WAVE format.
//
// Format precision must be 1 or 2 bytes.
func Encode(w io.WriteSeeker, s beep.Streamer, format beep.Format) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "wav")
		}
	}()

	if format.NumChannels <= 0 {
		return errors.New("wav: invalid number of channels (less than 1)")
	}
	if format.Precision != 1 && format.Precision != 2 {
		return errors.New("wav: unsupported precision, 1 or 2 is supported")
	}

	h := header{
		RiffMark:      [4]byte{'R', 'I', 'F', 'F'},
		FileSize:      -1, // finalization
		WaveMark:      [4]byte{'W', 'A', 'V', 'E'},
		FmtMark:       [4]byte{'f', 'm', 't', ' '},
		FormatSize:    16,
		FormatType:    1,
		NumChans:      int16(format.NumChannels),
		SampleRate:    int32(format.SampleRate),
		ByteRate:      int32(format.SampleRate * format.NumChannels * format.Precision),
		BytesPerFrame: int16(format.NumChannels * format.Precision),
		BitsPerSample: int16(format.Precision) * 8,
		DataMark:      [4]byte{'d', 'a', 't', 'a'},
		DataSize:      -1, // finalization
	}
	if err := binary.Write(w, binary.LittleEndian, &h); err != nil {
		return err
	}

	var (
		bw      = bufio.NewWriter(w)
		samples [512][2]float64
		written int
	)
	for {
		n, ok := s.Stream(samples[:])
		if !ok {
			break
		}
		switch {
		case format.Precision == 1 && format.NumChannels == 1:
			for _, sample := range samples[:n] {
				if err := encodeMono8(bw, sample); err != nil {
					return err
				}
			}
		case format.Precision == 1 && format.NumChannels >= 2:
			padding := make([]byte, (format.NumChannels-2)*format.Precision)
			for _, sample := range samples[:n] {
				if err := encodeStereo8(bw, sample); err != nil {
					return err
				}
				if _, err := bw.Write(padding); err != nil {
					return err
				}
			}
		case format.Precision == 2 && format.NumChannels == 1:
			for _, sample := range samples[:n] {
				if err := encodeMono16(bw, sample); err != nil {
					return err
				}
			}
		case format.Precision == 2 && format.NumChannels >= 2:
			padding := make([]byte, (format.NumChannels-2)*format.Precision)
			for _, sample := range samples[:n] {
				if err := encodeStereo16(bw, sample); err != nil {
					return err
				}
				if _, err := bw.Write(padding); err != nil {
					return err
				}
			}
		}
		written += n * format.NumChannels * format.Precision
	}
	if err := bw.Flush(); err != nil {
		return err
	}

	// finalize header
	h.FileSize = int32(44 + written) // 44 is the size of the header
	h.DataSize = int32(written)
	if _, err := w.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, &h); err != nil {
		return err
	}
	if _, err := w.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	return nil
}

func encodeMono8(w io.Writer, sample [2]float64) error {
	val := (sample[0] + sample[1]) / 2
	if val < -1 {
		val = -1
	}
	if val > +1 {
		val = +1
	}
	valUint8 := uint8((val + 1) / 2 * (1<<8 - 1))
	p := [1]byte{valUint8}
	_, err := w.Write(p[:])
	return err
}

func encodeMono16(w io.Writer, sample [2]float64) error {
	val := (sample[0] + sample[1]) / 2
	if val < -1 {
		val = -1
	}
	if val > +1 {
		val = +1
	}
	valInt16 := int16(val * (1<<15 - 1))
	low := byte(valInt16)
	high := byte(valInt16 >> 8)
	p := [2]byte{low, high}
	_, err := w.Write(p[:])
	return err
}

func encodeStereo8(w io.Writer, sample [2]float64) error {
	for _, val := range sample {
		if val < -1 {
			val = -1
		}
		if val > +1 {
			val = +1
		}
		valUint8 := uint8((val + 1) / 2 * (1<<8 - 1))
		p := [1]byte{valUint8}
		if _, err := w.Write(p[:]); err != nil {
			return err
		}
	}
	return nil
}

func encodeStereo16(w io.Writer, sample [2]float64) error {
	for _, val := range sample {
		if val < -1 {
			val = -1
		}
		if val > +1 {
			val = +1
		}
		valInt16 := int16(val * (1<<15 - 1))
		low := byte(valInt16)
		high := byte(valInt16 >> 8)
		p := [2]byte{low, high}
		if _, err := w.Write(p[:]); err != nil {
			return err
		}
	}
	return nil
}
