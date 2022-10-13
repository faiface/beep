// Package speaker implements playback of beep.Streamer values through physical speakers.
package speaker

import (
	"github.com/faiface/beep"
	"github.com/hajimehoshi/oto/v2"
	"github.com/pkg/errors"
	"io"
	"sync"
)

const channelCount = 2
const bitDepthInBytes = 2
const bytesPerSample = bitDepthInBytes * channelCount

var (
	mu      sync.Mutex
	mixer   beep.Mixer
	context *oto.Context
	player  oto.Player
)

// Init initializes audio playback through speaker. Must be called before using this package.
//
// The bufferSize argument specifies the number of samples of the speaker's buffer. Bigger
// bufferSize means lower CPU usage and more reliable playback. Lower bufferSize means better
// responsiveness and less delay.
func Init(sampleRate beep.SampleRate, bufferSize int) error {
	if context != nil {
		return errors.New("speaker cannot be initialized more than once")
	}

	mixer = beep.Mixer{}

	var err error
	var readyChan chan struct{}
	context, readyChan, err = oto.NewContext(int(sampleRate), channelCount, bitDepthInBytes)
	if err != nil {
		return errors.Wrap(err, "failed to initialize speaker")
	}
	<-readyChan

	player = context.NewPlayer(newReaderFromStreamer(&mixer))
	player.(oto.BufferSizeSetter).SetBufferSize(bufferSize * bytesPerSample)
	player.Play()

	return nil
}

// Close closes the playback and the driver. In most cases, there is certainly no need to call Close
// even when the program doesn't play anymore, because in properly set systems, the default mixer
// handles multiple concurrent processes. It's only when the default device is not a virtual but hardware
// device, that you'll probably want to manually manage the device from your application.
//
// TODO: investigate what happens now that oto.Context doesn't have a Close method.
func Close() {
	if player != nil {
		player.Close()
		player = nil
		mixer.Clear()
	}
}

// Lock locks the speaker. While locked, speaker won't pull new data from the playing Streamers. Lock
// if you want to modify any currently playing Streamers to avoid race conditions.
//
// Always lock speaker for as little time as possible, to avoid playback glitches.
func Lock() {
	mu.Lock()
}

// Unlock unlocks the speaker. Call after modifying any currently playing Streamer.
func Unlock() {
	mu.Unlock()
}

// Play starts playing all provided Streamers through the speaker.
func Play(s ...beep.Streamer) {
	mu.Lock()
	mixer.Add(s...)
	mu.Unlock()
}

// Clear removes all currently playing Streamers from the speaker.
func Clear() {
	mu.Lock()
	mixer.Clear()
	mu.Unlock()
}

// sampleReader is a wrapper for beep.Streamer to implement io.Reader.
type sampleReader struct {
	s   beep.Streamer
	buf [][2]float64
}

func newReaderFromStreamer(s beep.Streamer) *sampleReader {
	return &sampleReader{
		s: s,
	}
}

// Read pulls samples from the reader and fills buf with the encoded
// samples. Read expects the size of buf be divisible by the length
// of a sample (= channel count * bit depth in bytes).
func (s *sampleReader) Read(buf []byte) (n int, err error) {
	// Read samples from streamer
	if len(buf)%bytesPerSample != 0 {
		return 0, errors.New("requested number of bytes do not align with the samples")
	}
	ns := len(buf) / bytesPerSample
	if len(s.buf) < ns {
		s.buf = make([][2]float64, ns)
	}
	mu.Lock()
	ns, ok := s.s.Stream(s.buf[:ns])
	mu.Unlock()
	if !ok {
		if s.s.Err() != nil {
			return 0, errors.Wrap(s.s.Err(), "streamer returned error when requesting samples")
		}
		if ns == 0 {
			return 0, io.EOF
		}
	}

	// Convert samples to bytes
	for i := range s.buf[:ns] {
		for c := range s.buf[i] {
			val := s.buf[i][c]
			if val < -1 {
				val = -1
			}
			if val > +1 {
				val = +1
			}
			valInt16 := int16(val * (1<<15 - 1))
			low := byte(valInt16)
			high := byte(valInt16 >> 8)
			buf[i*bytesPerSample+c*bitDepthInBytes+0] = low
			buf[i*bytesPerSample+c*bitDepthInBytes+1] = high
		}
	}

	return ns * bytesPerSample, nil
}
