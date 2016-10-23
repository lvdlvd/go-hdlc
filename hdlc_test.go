package hdlc

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestThatItWorks(t *testing.T) {

	var buf bytes.Buffer
	framer := Frame(&buf)

	var slices [][]byte

	for n := 0; n < 2000; n++ {
		b := make([]byte, rand.Intn(10000))
		for i := range b {
			b[i] = byte(rand.Uint32())
		}
		slices = append(slices, b)

		if n, err := framer.Write(b); n != len(b) || err != nil {
			t.Errorf("writing to framer, expected %d, nil, got %d, %v", len(b), n, err)
		}

		// log.Printf("%x\n", b)
		// log.Printf("%x\n", buf.Bytes())

	}

	t.Logf("made %d slices, wrote %d bytes", len(slices), buf.Len())

	unframer := Unframe(bytes.NewReader(buf.Bytes()))
	if n, err := unframer.Resync(); n != 1 || err != nil {
		t.Errorf("resyncing expected 1, nil, got %d, %v", n, err)
	}

	for i, b := range slices {
		b2 := make([]byte, len(b))
		n, err := unframer.Read(b2)
		if err != nil {
			t.Errorf("reading slice %d from unframer, expected %d, nil, got %d, %v", i, len(b), n, err)
		}
		if bytes.Compare(b, b2) != 0 {
			t.Errorf("buffers differ")
		}
	}

}

// TODO
// func TestResync(t *testing.T)      {}
// func TestAbort(t *testing.T)       {}
// func TestTrailingEsc(t *testing.T) {}
