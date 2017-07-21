package beep

import "fmt"

// Format is the format of a Buffer or another audio source.
type Format struct {
	// SampleRate is the number of samples per second.
	SampleRate int

	// NumChannels is the number of channels. The value of 1 is mono, the value of 2 is stereo.
	// The samples should always be interleaved.
	NumChannels int

	// Precision is the number of bytes used to encode a single sample.
	Precision int
}

// Width returns the number of bytes per one sample (all channels).
//
// This is equal to f.NumChannels * f.Precision.
func (f Format) Width() int {
	return f.NumChannels * f.Precision
}

// EncodeSigned encodes a single sample in f.Width() bytes to p in signed format.
func (f Format) EncodeSigned(p []byte, sample [2]float64) (n int) {
	return f.encode(true, p, sample)
}

// EncodeUnsigned encodes a single sample in f.Width() bytes to p in unsigned format.
func (f Format) EncodeUnsigned(p []byte, sample [2]float64) (n int) {
	return f.encode(false, p, sample)
}

// DecodeSigned decodes a single sample encoded in f.Width() bytes from p in signed format.
func (f Format) DecodeSigned(p []byte) (sample [2]float64, n int) {
	return f.decode(true, p)
}

// DecodeUnsigned decodes a single sample encoded in f.Width() bytes from p in unsigned format.
func (f Format) DecodeUnsigned(p []byte) (sample [2]float64, n int) {
	return f.decode(false, p)
}

func (f Format) encode(signed bool, p []byte, sample [2]float64) (n int) {
	switch {
	case f.NumChannels == 1:
		x := norm((sample[0] + sample[1]) / 2)
		p = p[encodeFloat(signed, p, f.Precision, x):]
	case f.NumChannels >= 2:
		for c := range sample {
			x := norm(sample[c])
			p = p[encodeFloat(signed, p, f.Precision, x):]
		}
		for c := len(sample); c < f.NumChannels; c++ {
			p = p[encodeFloat(signed, p, f.Precision, 0):]
		}
	default:
		panic(fmt.Errorf("format: encode: invalid number of channels: %d", f.NumChannels))
	}
	return f.Width()
}

func (f Format) decode(signed bool, p []byte) (sample [2]float64, n int) {
	switch {
	case f.NumChannels == 1:
		x, _ := decodeFloat(signed, p, f.Precision)
		return [2]float64{x, x}, f.Width()
	case f.NumChannels >= 2:
		for c := range sample {
			x, n := decodeFloat(signed, p, f.Precision)
			sample[c] = x
			p = p[n:]
		}
		for c := len(sample); c < f.NumChannels; c++ {
			_, n := decodeFloat(signed, p, f.Precision)
			p = p[n:]
		}
		return sample, f.Width()
	default:
		panic(fmt.Errorf("format: decode: invalid number of channels: %d", f.NumChannels))
	}
}

func encodeFloat(signed bool, p []byte, precision int, x float64) (n int) {
	var xUint64 uint64
	if signed {
		xUint64 = floatToSigned(precision, x)
	} else {
		xUint64 = floatToUnsigned(precision, x)
	}
	for i := 0; i < precision; i++ {
		p[i] = byte(xUint64)
		xUint64 >>= 8
	}
	return precision
}

func decodeFloat(signed bool, p []byte, precision int) (x float64, n int) {
	var xUint64 uint64
	for i := 0; i < precision; i++ {
		xUint64 <<= 8
		xUint64 += uint64(p[i])
	}
	if signed {
		return signedToFloat(precision, xUint64), precision
	}
	return unsignedToFloat(precision, xUint64), precision
}

func floatToSigned(precision int, x float64) uint64 {
	return uint64(x * float64(uint64(1)<<uint(precision*8-1)-1))
}

func floatToUnsigned(precision int, x float64) uint64 {
	return uint64((x + 1) / 2 * float64(uint64(1)<<uint(precision*8)-1))
}

func signedToFloat(precision int, xUint64 uint64) float64 {
	return float64(int64(xUint64)) / float64(uint64(1)<<uint(precision*8-1)-1)
}

func unsignedToFloat(precision int, xUint64 uint64) float64 {
	return float64(xUint64)/float64(uint(1)<<uint(precision*8)-1)*2 - 1
}

func norm(x float64) float64 {
	if x < -1 {
		return -1
	}
	if x > +1 {
		return +1
	}
	return x
}
