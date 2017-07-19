package beep

// Format is the format of a Buffer or another audio source.
//
// N samples get encoded in N * NumChannels * Precision bytes.
type Format struct {
	// SampleRate is the number of samples per second.
	SampleRate int

	// NumChannels is the number of channels. The value of 1 is mono, the value of 2 is stereo.
	// The samples should always be interleaved.
	NumChannels int

	// Precision is the number of bytes used to encode a single sample.
	Precision int
}
