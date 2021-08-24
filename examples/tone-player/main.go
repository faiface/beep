package main

import "os"
import "fmt"
import "github.com/faiface/beep"
import "github.com/faiface/beep/speaker"
import "strconv"

func main() {
f,_ := strconv.Atoi(os.Args[1])
speaker.Init(beep.SampleRate(48000), 4800)
s := beep.SinTone(beep.SampleRate(48000), f)
rb := make([][2]float64,20)
s.Stream(rb)
fmt.Println(rb)
speaker.Play(s)
for {

}
}