//go:build malgo
// +build malgo

// Package speaker implements playback of beep.Streamer values through physical speakers.
package speaker

import (
	"fmt"
	"sync"

	"github.com/faiface/beep"
	"github.com/gen2brain/malgo"
	"github.com/pkg/errors"
)

var (
	mu      sync.Mutex
	mixer   beep.Mixer
	samples [][2]float64
	context *malgo.AllocatedContext
	player  *malgo.Device
	done    chan struct{}
	buf     []byte
)

type SpeakerDevice struct {
	info malgo.DeviceInfo
	Name string
}

type chooseDeviceCB func(deviceList []SpeakerDevice) *SpeakerDevice

// Init initializes audio playback through speaker. Must be called before using this package.
//
// The bufferSize argument specifies the number of samples of the speaker's buffer. Bigger
// bufferSize means lower CPU usage and more reliable playback. Lower bufferSize means better
// responsiveness and less delay.
func Init(sampleRate beep.SampleRate, bufferSize int) error {
	return InitDeviceSelection(sampleRate, bufferSize, nil)
}

func configure(sampleRate beep.SampleRate, cb chooseDeviceCB) (malgo.DeviceConfig, error) {

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = 2 //channels
	deviceConfig.SampleRate = uint32(sampleRate)
	deviceConfig.Alsa.NoMMap = 1

	if cb != nil {
		playbackDevices, err := context.Devices(malgo.Playback)
		if err != nil {
			return malgo.DeviceConfig{}, err
		}
		speakerList := []SpeakerDevice{}
		for _, device := range playbackDevices {
			speakerList = append(speakerList, SpeakerDevice{device, device.Name()})
		}
		ret := cb(speakerList)
		if ret != nil {
			deviceConfig.Playback.DeviceID = ret.info.ID.Pointer()
		}
	}
	return deviceConfig, nil
}

func InitDeviceSelection(sampleRate beep.SampleRate, bufferSize int, cb chooseDeviceCB) error {
	mu.Lock()
	defer mu.Unlock()

	Close()

	mixer = beep.Mixer{}

	samples = make([][2]float64, bufferSize)

	var err error
	context, err = malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("LOG <%v>\n", message)
	})

	if err != nil {
		return errors.Wrap(err, "failed to initialize speaker (context)")
	}

	var deviceConfig malgo.DeviceConfig
	deviceConfig, err = configure(sampleRate, cb)
	if err != nil {
		return errors.Wrap(err, "failed to initialize speaker (configure)")
	}

	onSamples := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		byteCount := framecount * deviceConfig.Playback.Channels * uint32(malgo.SampleSizeInBytes(deviceConfig.Playback.Format))
		if len(buf) < int(byteCount) {
			update()
		}
		copy(pOutputSample, buf[:byteCount])
		buf = append([]byte{}, buf[byteCount:]...)
	}

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onSamples,
	}
	player, err = malgo.InitDevice(context.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		return errors.Wrap(err, "failed to initialize speaker (player)")
	}

	err = player.Start()
	if err != nil {
		return errors.Wrap(err, "failed to initialize speaker (player start)")
	}

	done = make(chan struct{})

	go func() {
		for {
			select {
			default:
			case <-done:
				player.Stop()
				return
			}
		}
	}()

	return nil
}

// Close closes the playback and the driver. In most cases, there is certainly no need to call Close
// even when the program doesn't play anymore, because in properly set systems, the default mixer
// handles multiple concurrent processes. It's only when the default device is not a virtual but hardware
// device, that you'll probably want to manually manage the device from your application.
func Close() {
	if player != nil {
		if done != nil {
			done <- struct{}{}
			done = nil
		}
		player.Stop()
		player.Uninit()
		context.Uninit()
		player = nil
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

// update pulls new data from the playing Streamers and sends it to the speaker. Blocks until the
// data is sent and started playing.
func update() {
	mu.Lock()
	mixer.Stream(samples)
	mu.Unlock()

	for i := range samples {
		for c := range samples[i] {
			val := samples[i][c]
			if val < -1 {
				val = -1
			}
			if val > +1 {
				val = +1
			}
			valInt16 := int16(val * (1<<15 - 1))
			low := byte(valInt16)
			high := byte(valInt16 >> 8)
			buf = append(buf, low, high)
		}
	}
}
