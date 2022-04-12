package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s song.mp3\n", os.Args[0])
		os.Exit(1)
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		report(err)
	}
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		report(err)
	}
	defer streamer.Close()

	deviceSelect := func(deviceList []speaker.SpeakerDevice) *speaker.SpeakerDevice {
		for {
			fmt.Printf("choose a device from the list:\n")
			fmt.Println("0: default device")
			for i, d := range deviceList {
				fmt.Printf("%d: %s\n", i+1, d.Name)
			}

			var sindex string
			_, err := fmt.Scanf("%s\n", &sindex)
			if err != nil {
				fmt.Println("invalid selection, try again...")
				continue
			}
			if sindex == "q" {
				os.Exit(0)
			}
			index, _ := strconv.Atoi(sindex)
			if index > len(deviceList) {
				fmt.Printf("invalid selection [%d], try again...\n", index)
				continue
			}
			if index == 0 {
				return nil
			}
			return &deviceList[index-1]
		}
	}
	speaker.InitDeviceSelection(format.SampleRate, format.SampleRate.N(time.Second/10), deviceSelect)

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}

func report(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
