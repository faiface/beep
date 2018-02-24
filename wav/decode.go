package wav

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/faiface/beep"
	"github.com/pkg/errors"
)

// Decode takes a ReadCloser containing audio data in WAVE format and returns a StreamSeekCloser,
// which streams that audio. The Seek method will panic if rc is not io.Seeker.
//
// Do not close the supplied ReadSeekCloser, instead, use the Close method of the returned
// StreamSeekCloser when you want to release the resources.
func Decode(rc io.ReadCloser) (s beep.StreamSeekCloser, format beep.Format, err error) {
	d := decoder{rc: rc}
	defer func() { // hacky way to always close rsc if an error occured
		if err != nil {
			d.rc.Close()
		}
	}()
	if err := binary.Read(rc, binary.LittleEndian, &d.h); err != nil {
		return nil, beep.Format{}, errors.Wrap(err, "wav")
	}
	if string(d.h.RiffMark[:]) != "RIFF" {
		return nil, beep.Format{}, errors.New("wav: missing RIFF at the beginning")
	}
	if string(d.h.WaveMark[:]) != "WAVE" {
		return nil, beep.Format{}, errors.New("wav: unsupported file type")
	}
	if string(d.h.FmtMark[:]) != "fmt " {
		return nil, beep.Format{}, errors.New("wav: missing format chunk marker")
	}
	if string(d.h.DataMark[:]) != "data" {
		return nil, beep.Format{}, errors.New("wav: missing data chunk marker")
	}
	if d.h.FormatType != 1 {
		return nil, beep.Format{}, errors.New("wav: unsupported format type")
	}
	if d.h.NumChans <= 0 {
		return nil, beep.Format{}, errors.New("wav: invalid number of channels (less than 1)")
	}
	if d.h.BitsPerSample != 8 && d.h.BitsPerSample != 16 {
		return nil, beep.Format{}, errors.New("wav: unsupported number of bits per sample, 8 or 16 are supported")
	}
	format = beep.Format{
		SampleRate:  beep.SampleRate(d.h.SampleRate),
		NumChannels: int(d.h.NumChans),
		Precision:   int(d.h.BitsPerSample / 8),
	}
	return &d, format, nil
}

type header struct {
	RiffMark      [4]byte
	FileSize      int32
	WaveMark      [4]byte
	FmtMark       [4]byte
	FormatSize    int32
	FormatType    int16
	NumChans      int16
	SampleRate    int32
	ByteRate      int32
	BytesPerFrame int16
	BitsPerSample int16
	DataMark      [4]byte
	DataSize      int32
}

type decoder struct {
	rc  io.ReadCloser
	h   header
	pos int32
	err error
}

func (d *decoder) Stream(samples [][2]float64) (n int, ok bool) {
	if d.err != nil || d.pos >= d.h.DataSize {
		return 0, false
	}
	bytesPerFrame := int(d.h.BytesPerFrame)
	numBytes := int32(len(samples) * bytesPerFrame)
	if numBytes > d.h.DataSize-d.pos {
		numBytes = d.h.DataSize - d.pos
	}
	p := make([]byte, numBytes)
	n, err := d.rc.Read(p)
	if err != nil && err != io.EOF {
		d.err = err
	}
	switch {
	case d.h.BitsPerSample == 8 && d.h.NumChans == 1:
		for i, j := 0, 0; i <= n-bytesPerFrame; i, j = i+bytesPerFrame, j+1 {
			val := float64(p[i])/(1<<8-1)*2 - 1
			samples[j][0] = val
			samples[j][1] = val
		}
	case d.h.BitsPerSample == 8 && d.h.NumChans >= 2:
		for i, j := 0, 0; i <= n-bytesPerFrame; i, j = i+bytesPerFrame, j+1 {
			samples[j][0] = float64(p[i+0])/(1<<8-1)*2 - 1
			samples[j][1] = float64(p[i+1])/(1<<8-1)*2 - 1
		}
	case d.h.BitsPerSample == 16 && d.h.NumChans == 1:
		for i, j := 0, 0; i <= n-bytesPerFrame; i, j = i+bytesPerFrame, j+1 {
			val := float64(int16(p[i+0])+int16(p[i+1])*(1<<8)) / (1<<15 - 1)
			samples[j][0] = val
			samples[j][1] = val
		}
	case d.h.BitsPerSample == 16 && d.h.NumChans >= 2:
		for i, j := 0, 0; i <= n-bytesPerFrame; i, j = i+bytesPerFrame, j+1 {
			samples[j][0] = float64(int16(p[i+0])+int16(p[i+1])*(1<<8)) / (1<<15 - 1)
			samples[j][1] = float64(int16(p[i+2])+int16(p[i+3])*(1<<8)) / (1<<15 - 1)
		}
	}
	d.pos += int32(n)
	return n / bytesPerFrame, true
}

func (d *decoder) Err() error {
	return d.err
}

func (d *decoder) Len() int {
	numBytes := time.Duration(d.h.DataSize)
	perFrame := time.Duration(d.h.BytesPerFrame)
	return int(numBytes / perFrame)
}

func (d *decoder) Position() int {
	return int(d.pos / int32(d.h.BytesPerFrame))
}

func (d *decoder) Seek(p int) error {
	seeker, ok := d.rc.(io.Seeker)
	if !ok {
		panic(fmt.Errorf("wav: seek: resource is not io.Seeker"))
	}
	if p < 0 || d.Len() < p {
		return fmt.Errorf("wav: seek position %v out of range [%v, %v]", p, 0, d.Len())
	}
	pos := int32(p) * int32(d.h.BytesPerFrame)
	_, err := seeker.Seek(int64(pos)+44, io.SeekStart) // 44 is the size of the header
	if err != nil {
		return errors.Wrap(err, "wav: seek error")
	}
	d.pos = pos
	return nil
}

func (d *decoder) Close() error {
	err := d.rc.Close()
	if err != nil {
		return errors.Wrap(err, "wav")
	}
	return nil
}
