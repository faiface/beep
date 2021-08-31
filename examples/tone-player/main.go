package main

import "os"
import "github.com/faiface/beep/generators"
import "github.com/faiface/beep/speaker"
import "github.com/faiface/beep"
import "strconv"

func main() {
f,_ := strconv.Atoi(os.Args[1])
speaker.Init(beep.SampleRate(48000), 4800)
s := generators.SinTone(beep.SampleRate(48000), f)
speaker.Play(s)
for {

}
}