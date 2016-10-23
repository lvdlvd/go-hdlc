package hdlc

import (
	"bufio"
	"errors"
	"io"
)

var (
	ErrResynced = errors.New("resynced to hdlc flag")
	ErrAbort    = errors.New("aborted hdlc frame")
)

const bufSize = 8192 // we need to know for calls to peek, can't rely on bufio internal

// Unframe returns a new Unframer that reads from r.
func Unframe(r io.Reader) *Unframer { return &Unframer{r: bufio.NewReaderSize(r, bufSize)} }

// An unframer is an io.Reader that returns one frame per call to Read.
//
// For receiving partial frames and resyncing, use Resync and ReadUnescaped.
type Unframer struct {
	r           *bufio.Reader
	lastwasflag bool
}

// Read makes Unframer satisfy io.Read.
//
// It tries to return one complete unescaped packet.
// It returns any error Resync or ReadUnescaped may have
// returned.
// If err == ErrResynced, n is the number of bytes discarded
// including flags, and p will be unmodified.
func (u *Unframer) Read(p []byte) (n int, err error) {
	n, err = u.Resync()
	if n > 0 && err == nil {
		err = ErrResynced
	}
	if err != nil {
		return
	}
	return u.ReadEscaped(p)
}

// Resync discards bytes until the last we read was FLAG and the next is not a FLAG.
// returns the number of bytes discarded, including the flags, and any error returned
// from the underlying reader before finding the end of the flags.
func (u *Unframer) Resync() (n int, err error) {
	var buf []byte
	for !u.lastwasflag {
		buf, err = u.r.ReadSlice(FLAG)
		n += len(buf)
		if err != bufio.ErrBufferFull {
			break
		}
	}

	for err == nil {
		u.lastwasflag = true

		buf, err = u.r.Peek(bufSize)
		if err == io.EOF {
			err = nil
		}
		idx := findNot(buf, FLAG)
		if idx == 0 {
			break
		}

		var nn int
		nn, err = u.r.Discard(idx)
		n += nn
	}

	return
}

// findNot returns the index of the first byte that is NOT equal to c
// or len(p) if p consists only of copies of c.
func findNot(p []byte, c byte) int {
	for i, v := range p {
		if v != c {
			return i
		}
	}
	return len(p)
}

// ReadEscaped reads up to len(p) bytes from the pending frame,
// unescaping the contents.
//
// In all cases n is the number of bytes read into p, and err
// is set as follows:
// nil:  a frame was read till completion
// bufio.ErrBufferFull: reading was succesful but p was not large enough
// to read the full frame. A next call to ReadUnescaped
// will return the next part.
// io.ErrUnexpectedEOF means the underlying reader returned an EOF
// before we found the end-of-frame flag.
// ErrAbort: the underlying bytestream contained an ABORT.
// all other values, including io.EOF come from the underlying reader.
//
func (u *Unframer) ReadEscaped(p []byte) (n int, err error) {
	isEsc := false
	for {
		var val byte
		val, err = u.r.ReadByte()
		if err == io.EOF {
			return n, io.ErrUnexpectedEOF
		}
		if err != nil {
			return n, err
		}

		if val == FLAG {
			u.lastwasflag = true
			return n, nil
		}
		u.lastwasflag = false
		if val == ABORT {
			return n, ErrAbort
		}

		if n >= len(p) {
			// we have an esc or a value but we won't
			// have room to write it, put it back for the next call
			if err := u.r.UnreadByte(); err != nil {
				return n, err
			}
			return n, bufio.ErrBufferFull
		}

		if val == ESC {
			isEsc = true
			continue
		}
		if isEsc {
			isEsc = false
			val ^= 0x20
		}
		p[n] = val
		n++
	}
	panic("not reached")
}
