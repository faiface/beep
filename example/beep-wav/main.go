package main

import (
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

func main() {
	input := "sound.wav"
	f, errOpen := os.Open(input)
	if errOpen != nil {
		log.Fatalf("open: %s: %s", input, errOpen)
	}

	s, format, errDec := wav.Decode(f)
	if errDec != nil {
		log.Fatalf("decode: %s", errDec)
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan struct{})

	log.Printf("playing")

	speaker.Play(beep.Seq(s, beep.Callback(func() {
		close(done)
	})))

	<-done

	log.Printf("done")
}
