package main

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/generators"
	"github.com/faiface/beep/speaker"
	"os"
	"strconv"
)

func usage() {
	fmt.Printf("usage: %s freq\n", os.Args[0])
	fmt.Println("where freq must be an integer from 1 to 24000")
	fmt.Println("24000 because samplerate of 48000 is hardcoded")
}
func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}
	f, err := strconv.Atoi(os.Args[1])
	if err != nil {
		usage()
		return
	}
	speaker.Init(beep.SampleRate(48000), 4800)
	s, err := generators.SinTone(beep.SampleRate(48000), f)
	if err != nil {
		panic(err)
	}
	speaker.Play(s)
	for {

	}
}
