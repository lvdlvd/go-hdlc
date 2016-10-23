package hdlc

import (
	"io"
	"math/rand"
	"testing"
)

func TestThatItWorks(t *testing.T) {

	r, w := io.Pipe()
	unframer, framer := Unframe(r), Frame(w)

	ch := make(chan []byte)

	// write 1000 random slices to the framer and the channel
	go func() {
		defer close(ch)
		for n := 0; n < 1000; n++ {
			b := make([]byte, rand.Int(20000))
			for i := range b {
				b[i] = byte(rand.Uint32())
			}

			ch <- b
			if n, err := framer.Write(b); n != len(b) || err == nil {
				t.Errorf("writing to framer, expected %d, nil, got %d, %v", len(b), n, err)
				return
			}
		}
	}()

	for b := range ch {
		b2 := make([]byte, len(b))

	}

}
