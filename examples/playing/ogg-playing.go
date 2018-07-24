package main

import (
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
)

func main() {
	// Open first sample File
	f, err := os.Open("sample1.ogg")

	// Check for errors when opening the file
	if err != nil {
		log.Fatal(err)
	}

	// Decode the .ogg File, if you have a .wav file, use wav.Decode(f)
	s, format, err := vorbis.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	// Init the Speaker with the SampleRate of the format and a buffer size of 1/10s
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// Channel, which will signal the end of the playback.
	playing := make(chan struct{})

	// Now we Play our Streamer on the Speaker
	speaker.Play(beep.Seq(s, beep.Callback(func() {
		// Callback after the stream Ends
		close(playing)
	})))
	<-playing
}
