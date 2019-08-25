package beep_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/faiface/beep"
)

func TestFormatEncodeDecode(t *testing.T) {
	formats := make(chan beep.Format)
	go func() {
		defer close(formats)
		for _, sampleRate := range []beep.SampleRate{100, 2347, 44100, 48000} {
			for _, numChannels := range []int{1, 2, 3, 4} {
				for _, precision := range []int{1, 2, 3, 4, 5, 6} {
					formats <- beep.Format{
						SampleRate:  sampleRate,
						NumChannels: numChannels,
						Precision:   precision,
					}
				}
			}
		}
	}()

	for format := range formats {
		for i := 0; i < 20; i++ {
			deviation := 2.0 / (math.Pow(2, float64(format.Precision)*8) - 2)
			sample := [2]float64{rand.Float64()*2 - 1, rand.Float64()*2 - 1}

			tmp := make([]byte, format.Width())
			format.EncodeSigned(tmp, sample)
			decoded, _ := format.DecodeSigned(tmp)

			if format.NumChannels == 1 {
				if math.Abs((sample[0]+sample[1])/2-decoded[0]) > deviation || decoded[0] != decoded[1] {
					t.Fatalf("signed decoded sample is too different: %v -> %v (deviation: %v)", sample, decoded, deviation)
				}
			} else {
				if math.Abs(sample[0]-decoded[0]) > deviation || math.Abs(sample[1]-decoded[1]) > deviation {
					t.Fatalf("signed decoded sample is too different: %v -> %v (deviation: %v)", sample, decoded, deviation)
				}
			}

			format.EncodeUnsigned(tmp, sample)
			decoded, _ = format.DecodeUnsigned(tmp)

			if format.NumChannels == 1 {
				if math.Abs((sample[0]+sample[1])/2-decoded[0]) > deviation || decoded[0] != decoded[1] {
					t.Fatalf("unsigned decoded sample is too different: %v -> %v (deviation: %v)", sample, decoded, deviation)
				}
			} else {
				if math.Abs(sample[0]-decoded[0]) > deviation || math.Abs(sample[1]-decoded[1]) > deviation {
					t.Fatalf("unsigned decoded sample is too different: %v -> %v (deviation: %v)", sample, decoded, deviation)
				}
			}
		}
	}
}

func TestBufferAppendPop(t *testing.T) {
	formats := make(chan beep.Format)
	go func() {
		defer close(formats)
		for _, numChannels := range []int{1, 2, 3, 4} {
			formats <- beep.Format{
				SampleRate:  44100,
				NumChannels: numChannels,
				Precision:   2,
			}
		}
	}()

	for format := range formats {
		b := beep.NewBuffer(format)
		b.Append(beep.Silence(768))
		if b.Len() != 768 {
			t.Fatalf("buffer length isn't equal to appended stream length: expected: %v, actual: %v (NumChannels: %v)", 768, b.Len(), format.NumChannels)
		}
		b.Pop(512)
		if b.Len() != 768-512 {
			t.Fatalf("buffer length isn't as expected after Pop: expected: %v, actual: %v (NumChannels: %v)", 768-512, b.Len(), format.NumChannels)
		}
	}
}
