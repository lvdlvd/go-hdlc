package hdlc

import (
	"bytes"
	"io"
)

// Framer is an io.Writer that frames each
// call to Write() into 1 frame.
//
// For sending partial and abortable frames
// use Flag, WriteEscaped and Abort.
type Framer struct {
	w           io.Writer
	lastwasflag bool
}

// Frame returns a new framer that writes to w.
func Frame(w io.Writer) *Framer { return &Framer{w: w} }

// Write adds framing information and escapes the payload.
func (f *Framer) Write(p []byte) (n int, err error) {
	if err = f.Flag(); err != nil {
		return
	}
	if n, err = f.WriteEscaped(p); err != nil {
		return
	}
	err = f.Flag()
	return
}

// Flag writes a 0x7e to the underlying writer, but only
// if the previous byte wasn't one.
// This is required to start and to end a frame filled by WriteEscaped.
// Repeated calls are suppressed.
func (f *Framer) Flag() error {
	if f.lastwasflag {
		return nil
	}
	n, err := f.w.Write([]byte{FLAG})
	if n > 0 {
		f.lastwasflag = true
	}
	return err
}

// Abort sends an abort sequence but only if the last byte wasnt a flag.
func (f *Framer) Abort() error {
	if f.lastwasflag {
		return nil
	}
	_, err := f.w.Write([]byte{ABORT})
	return err
}

// WriteEscaped writes p to f's underlying stream escaping all occurences of
// the bytes 0x7d, 0x7e and 0x7f.  It returns the number of bytes of p
// that were successfully encoded and any error returned by the underlying writer.
func (f *Framer) WriteEscaped(p []byte) (n int, err error) {
	defer func() {
		if n > 0 {
			f.lastwasflag = false
		}
	}()

	idxEsc, idxFlag, idxAbort := find(p, ESC), find(p, FLAG), find(p, ABORT)
	var nn int
	for len(p) > 0 {
		idx := min3(idxEsc, idxFlag, idxAbort)
		if idx > 0 {
			nn, err = f.w.Write(p[:idx])
			n += nn
			if err != nil {
				return
			}
		}
		if idx == len(p) {
			break
		}

		// if we're here p[0] should be one of the control characters and we should escape it
		nn, err = f.w.Write([]byte{ESC, p[idx] ^ 0x20})
		if err != nil {
			if nn > 0 {
				// if this is our first value, and the escape was written but not the escaped
				// value, we should still clear lastwasflag
				f.lastwasflag = false
			}
			return
		}
		n++
		idx++

		p = p[idx:]
		idxEsc -= idx
		idxFlag -= idx
		idxAbort -= idx

		if idxEsc < 0 {
			idxEsc = find(p, ESC)
		}
		if idxFlag < 0 {
			idxFlag = find(p, FLAG)
		}
		if idxAbort < 0 {
			idxAbort = find(p, ABORT)
		}
	}
	return
}

// find fixes indexbyte so it returns len(p) instead of -1 when not found.
func find(p []byte, c byte) int {
	if n := bytes.IndexByte(p, c); n >= 0 {
		return n
	}
	return len(p)
}

func min3(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= a && b <= c {
		return b
	}
	if c <= a && c <= b {
		return c
	}
	panic("should not happen")
}
