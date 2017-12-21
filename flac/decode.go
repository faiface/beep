package flac

import (
	"fmt"
	"io"

	"github.com/faiface/beep"
	"github.com/mewkiz/flac"
	"github.com/pkg/errors"
)

// Decode takes a ReadCloser containing audio data in FLAC format and returns a StreamSeekCloser,
// which streams that audio. The Seek method will panic if rc is not io.Seeker.
//
// Do not close the supplied ReadSeekCloser, instead, use the Close method of the returned
// StreamSeekCloser when you want to release the resources.
func Decode(rc io.ReadCloser) (s beep.StreamSeekCloser, format beep.Format, err error) {
	d := decoder{rc: rc}
	defer func() { // hacky way to always close rc if an error occured
		if err != nil {
			d.rc.Close()
		}
	}()
	rsc, ok := rc.(io.ReadSeeker)
	if !ok {
		panic(fmt.Errorf("%T does not implement io.Seeker", rc))
	}
	d.stream, err = flac.New(rsc)
	if err != nil {
		return nil, beep.Format{}, errors.Wrap(err, "flac")
	}
	format = beep.Format{
		SampleRate:  beep.SampleRate(d.stream.Info.SampleRate),
		NumChannels: int(d.stream.Info.NChannels),
		Precision:   int(d.stream.Info.BitsPerSample / 8),
	}
	return &d, format, nil
}

type decoder struct {
	rc     io.ReadCloser
	stream *flac.Stream
	buf    [][2]float64
	err    error
}

func (d *decoder) Stream(samples [][2]float64) (n int, ok bool) {
	if d.err != nil {
		return 0, false
	}
	// Copy samples from buffer.
	j := 0
	for i := range samples {
		if j >= len(d.buf) {
			// refill buffer.
			if err := d.refill(); err != nil {
				d.err = err
				return n, n > 0
			}
			j = 0
		}
		samples[i] = d.buf[j]
		j++
		n++
	}
	d.buf = d.buf[j:]
	return n, true
}

// refill decodes audio samples to fill the decode buffer.
func (d *decoder) refill() error {
	// Empty buffer.
	d.buf = d.buf[:0]
	// Parse audio frame.
	frame, err := d.stream.ParseNext()
	if err != nil {
		return err
	}
	// Expand buffer size if needed.
	n := len(frame.Subframes[0].Samples)
	if cap(d.buf) < n {
		d.buf = make([][2]float64, n)
	} else {
		d.buf = d.buf[:n]
	}
	// Decode audio samples.
	bps := d.stream.Info.BitsPerSample
	nchannels := d.stream.Info.NChannels
	switch {
	case bps == 8 && nchannels == 1:
		for i := 0; i < n; i++ {
			d.buf[i][0] = float64(int8(frame.Subframes[0].Samples[i])) / (1<<7 - 1)
			d.buf[i][1] = float64(int8(frame.Subframes[0].Samples[i])) / (1<<7 - 1)
		}
	case bps == 16 && nchannels == 1:
		for i := 0; i < n; i++ {
			d.buf[i][0] = float64(int16(frame.Subframes[0].Samples[i])) / (1<<15 - 1)
			d.buf[i][1] = float64(int16(frame.Subframes[0].Samples[i])) / (1<<15 - 1)
		}
	case bps == 24 && nchannels == 1:
		for i := 0; i < n; i++ {
			d.buf[i][0] = float64(int32(frame.Subframes[0].Samples[i])) / (1<<23 - 1)
			d.buf[i][1] = float64(int32(frame.Subframes[1].Samples[i])) / (1<<23 - 1)
		}
	case bps == 8 && nchannels >= 2:
		for i := 0; i < n; i++ {
			d.buf[i][0] = float64(int8(frame.Subframes[0].Samples[i])) / (1<<7 - 1)
			d.buf[i][1] = float64(int8(frame.Subframes[1].Samples[i])) / (1<<7 - 1)
		}
	case bps == 16 && nchannels >= 2:
		for i := 0; i < n; i++ {
			d.buf[i][0] = float64(int16(frame.Subframes[0].Samples[i])) / (1<<15 - 1)
			d.buf[i][1] = float64(int16(frame.Subframes[1].Samples[i])) / (1<<15 - 1)
		}
	case bps == 24 && nchannels >= 2:
		for i := 0; i < n; i++ {
			d.buf[i][0] = float64(int32(frame.Subframes[0].Samples[i])) / (1<<23 - 1)
			d.buf[i][1] = float64(int32(frame.Subframes[1].Samples[i])) / (1<<23 - 1)
		}
	default:
		panic(fmt.Errorf("support for %d bits-per-sample and %d channels combination not yet implemented", bps, nchannels))
	}
	return nil
}

func (d *decoder) Err() error {
	return d.err
}

func (d *decoder) Len() int {
	panic("not yet implemented")
}

func (d *decoder) Position() int {
	panic("not yet implemented")
}

func (d *decoder) Seek(p int) error {
	panic("not yet implemented")
}

func (d *decoder) Close() error {
	err := d.rc.Close()
	if err != nil {
		return errors.Wrap(err, "flac")
	}
	return nil
}
