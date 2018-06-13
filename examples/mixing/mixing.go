package main

import (
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

func main() {
	// Open and decode both Files
	f1, _ := os.Open("mix-1.mp3") // the guitar track
	s1, format, _ := mp3.Decode(f1)
	f2, _ := os.Open("mix-2.mp3") // the electro sample
	s2, format, _ := mp3.Decode(f2)

	// Create an beep.Mixer and add the two Streamers to it
	mixer := new(beep.Mixer)
	mixer.Play(s1)
	mixer.Play(s2)

	// Init the Speaker with the SampleRate of the format and a buffer size of 1/10s
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// Channel, which will signal the end of the playback.
	playing := make(chan struct{})

	// Now we Play our Mixer on the Speaker
	speaker.Play(beep.Seq(mixer, beep.Callback(func() {
		// Callback after the stream Ends
		close(playing)
	})))
	<-playing
}
