package main

import (
	"fmt"
	"math"
	"os"
	"time"
	"unicode"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/gdamore/tcell"
)

func multiplyChannels(left, right float64, s beep.Streamer) beep.Streamer {
	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		n, ok = s.Stream(samples)
		for i := range samples[:n] {
			samples[i][0] *= left
			samples[i][1] *= right
		}
		return n, ok
	})
}

type movingStreamer struct {
	x, y         float64
	velX, velY   float64
	leftDoppler  beep.Streamer
	rightDoppler beep.Streamer
}

func newMovingStreamer(sr beep.SampleRate, x, y float64, streamer beep.Streamer) *movingStreamer {
	ms := &movingStreamer{x: x, y: y}

	const metersPerSecond = 343
	samplesPerSecond := float64(sr)
	samplesPerMeter := samplesPerSecond / metersPerSecond

	leftEar, rightEar := beep.Dup(streamer)
	leftEar = multiplyChannels(1, 0, leftEar)
	rightEar = multiplyChannels(0, 1, rightEar)

	const earDistance = 0.16
	ms.leftDoppler = effects.Doppler(2, samplesPerMeter, leftEar, func(delta int) float64 {
		dt := sr.D(delta).Seconds()
		ms.x += ms.velX * dt
		ms.y += ms.velY * dt
		return math.Max(0.25, math.Hypot(ms.x+earDistance/2, ms.y))
	})
	ms.rightDoppler = effects.Doppler(2, samplesPerMeter, rightEar, func(delta int) float64 {
		return math.Max(0.25, math.Hypot(ms.x-earDistance/2, ms.y))
	})

	return ms
}

func (ms *movingStreamer) play() {
	speaker.Play(ms.leftDoppler, ms.rightDoppler)
}

func drawCircle(screen tcell.Screen, x, y float64, style tcell.Style) {
	width, height := screen.Size()
	centerX, centerY := float64(width)/2, float64(height)/2

	lx, ly := int(centerX+(x-0.25)*2), int(centerY+y)
	screen.SetContent(lx, ly, tcell.RuneBlock, nil, style)

	rx, ry := int(centerX+(x+0.25)*2), int(centerY+y)
	screen.SetContent(rx, ry, tcell.RuneBlock, nil, style)
}

func drawTextLine(screen tcell.Screen, x, y int, s string, style tcell.Style) {
	for _, r := range s {
		screen.SetContent(x, y, r, nil, style)
		x++
	}
}

func drawHelp(screen tcell.Screen, style tcell.Style) {
	drawTextLine(screen, 0, 0, "Welcome to the Doppler Stereo Room!", style)
	drawTextLine(screen, 0, 1, "Press [ESC] to quit.", style)

	drawTextLine(screen, 0, 2, "Move the", style)
	drawTextLine(screen, 9, 2, "LEFT", style.Background(tcell.ColorGreen).Foreground(tcell.ColorWhiteSmoke))
	drawTextLine(screen, 14, 2, "speaker with WASD.", style)

	drawTextLine(screen, 0, 3, "Move the", style)
	drawTextLine(screen, 9, 3, "RIGHT", style.Background(tcell.ColorBlue).Foreground(tcell.ColorWhiteSmoke))
	drawTextLine(screen, 15, 3, "speaker with IJKL.", style)

	drawTextLine(screen, 0, 4, "Move the", style)
	drawTextLine(screen, 9, 4, "BOTH", style.Background(tcell.ColorDeepPink).Foreground(tcell.ColorWhiteSmoke))
	drawTextLine(screen, 15, 4, "speakers with the Numpad Buttons 1-9.", style)

	drawTextLine(screen, 0, 5, "Press to start moving, press again to stop. Use [SHIFT] to move fast.", style)
}

type DirectionMode int

const (
	_ DirectionMode = iota
	Applied
	SetPoint
)

type EventMappedLocation struct {
	lx,
	ly,
	rx,
	ry float64
	using DirectionMode
}

var ResetLocation = EventMappedLocation{-1, 0, 1, 0, SetPoint}

var directions = map[rune]EventMappedLocation{

	// Reset
	'5': ResetLocation,
	'r': ResetLocation,

	// Numb Pad Layout Mapped
	// Front, Back
	'8': {-1, -1, 1, -1, SetPoint},
	'2': {-1, 1, 1, 1, SetPoint},
	// Left, Right
	'4': {-1.5, 0, -1, 0, SetPoint},
	'6': { 1,0,  1.5,0, SetPoint},
	// Layout Top Left, Top Right, Bottom Right, Bottom Left
	'7': {-1.8, -1.8,-0.8, -1.8, SetPoint},
	'9': {0.8, -1.8,1.8, -1.8,  SetPoint},
	'1': {-1.8, 1.8, -0.8, 1.8, SetPoint},
	'3': {0.8, 1.8, 1.8, 1.8, SetPoint},

	// Diagonal Locations
	'\\': {-1, -1, 1, 1, SetPoint},
	'/': {-1, 1, 1, -1, SetPoint},

	// Left
	'a': {-1, 0, 0, 0, Applied},
	'd': {+1, 0, 0, 0, Applied},
	'w': {0, -1, 0, 0, Applied},
	's': {0, +1, 0, 0, Applied},

	// Right
	'j': {0, 0, -1, 0, Applied},
	'l': {0, 0, +1, 0, Applied},
	'i': {0, 0, 0, -1, Applied},
	'k': {0, 0, 0, +1, Applied},
}

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

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/30))

	leftCh, rightCh := beep.Dup(streamer)

	leftCh = effects.Mono(multiplyChannels(1, 0, leftCh))
	rightCh = effects.Mono(multiplyChannels(0, 1, rightCh))

	leftMS := newMovingStreamer(format.SampleRate, -1, 0, leftCh)
	rightMS := newMovingStreamer(format.SampleRate, +1, 0, rightCh)

	leftMS.play()
	rightMS.play()

	screen, err := tcell.NewScreen()
	if err != nil {
		report(err)
	}
	err = screen.Init()
	if err != nil {
		report(err)
	}
	defer screen.Fini()

	frames := time.Tick(time.Second / 30)
	events := make(chan tcell.Event)
	go func() {
		for {
			events <- screen.PollEvent()
		}
	}()

loop:
	for {
		select {
		case <-frames:
			speaker.Lock()
			lx, ly := leftMS.x, leftMS.y
			rx, ry := rightMS.x, rightMS.y
			speaker.Unlock()

			style := tcell.StyleDefault.
				Background(tcell.ColorWhiteSmoke).
				Foreground(tcell.ColorBlack)

			screen.Clear()
			screen.Fill(' ', style)
			drawHelp(screen, style)
			drawCircle(screen, 0, 0, style.Foreground(tcell.ColorBlack))
			drawCircle(screen, lx*2, ly*2, style.Foreground(tcell.ColorGreen))
			drawCircle(screen, rx*2, ry*2, style.Foreground(tcell.ColorBlue))
			screen.Show()

		case event := <-events:
			switch event := event.(type) {
			case *tcell.EventKey:
				if event.Key() == tcell.KeyESC {
					break loop
				}

				if event.Key() != tcell.KeyRune {
					break
				}

				const (
					slowSpeed = 2.0
					fastSpeed = 16.0
				)

				speaker.Lock()

				speed := slowSpeed
				if unicode.ToLower(event.Rune()) != event.Rune() {
					speed = fastSpeed
				}

				dir := directions[unicode.ToLower(event.Rune())]

				if dir.using == Applied {
					if dir.lx != 0 {
						if leftMS.velX == dir.lx*speed {
							leftMS.velX = 0
						} else {
							leftMS.velX = dir.lx * speed
						}
					}
					if dir.ly != 0 {
						if leftMS.velY == dir.ly*speed {
							leftMS.velY = 0
						} else {
							leftMS.velY = dir.ly * speed
						}
					}
					if dir.rx != 0 {
						if rightMS.velX == dir.rx*speed {
							rightMS.velX = 0
						} else {
							rightMS.velX = dir.rx * speed
						}
					}
					if dir.ry != 0 {
						if rightMS.velY == dir.ry*speed {
							rightMS.velY = 0
						} else {
							rightMS.velY = dir.ry * speed
						}
					}
				} else if dir.using == SetPoint {
					leftMS.velX = 0
					leftMS.velY = 0
					rightMS.velX = 0
					rightMS.velY = 0

					leftMS.x = dir.lx
					leftMS.y = dir.ly
					rightMS.x = dir.rx
					rightMS.y = dir.ry
				}

				speaker.Unlock()
			}
		}
	}
}

func report(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
