package hdlc

import (
	"bufio"
	"errors"
	"io"
)

var (
	ErrResynced = errors.New("resynced to hdlc flag")
	ErrPartial  = errors.New("partial hdlc frame")
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
	return u.ReadUnescaped(p)
}

// Resync discards bytes until the last we read was FLAG and the next is not a FLAG.
// returns the number of bytes discarded, including the flags, and any error returned
// from the underlying reader before finding the end of the flags.
func (u *Unframer) Resync() (n int, err error) {
	if u.lastwasflag {
		return 0, nil
	}

	var buf []byte
	for {
		buf, err = u.r.Peek(bufSize)
		if err != nil && err != io.EOF {
			return
		}

		if len(buf) == 0 {
			return // will carry the eof from above
		}

		idx := find(buf, FLAG)
		if idx == 0 {
			break
		}

		var nn int
		nn, err = u.r.Discard(idx)
		n += nn
		if err != nil {
			return
		}
	}

	if len(buf) == 0 || buf[0] != FLAG {
		panic("resyncing lost the flag")
	}

	for len(buf) > 0 && buf[0] == FLAG {
		idx := findNot(buf, FLAG)
		var nn int
		nn, err = u.r.Discard(idx)
		if nn > 0 {
			u.lastwasflag = true
		}
		n += nn
		if err != nil {
			return
		}

		buf, err = u.r.Peek(bufSize)
		if err != nil && err != io.EOF {
			return
		}
	}

	return
}

// ReadUnescaped reads up to len(p) bytes from the pending frame,
// unescaping the contents.
//
// In all cases n is the number of bytes read into p, and err
// is set as follows:
// nil:  a frame was read till completion
// ErrPartial: reading was succesful but p was not large enough
// to read the full frame. A next call to ReadUnescaped
// will return the next part.
// io.ErrUnexpectedEOF means the underlying reader returned an EOF
// before we found the end-of-frame flag.
// ErrAbort: the underlying bytestream contained an ABORT.
// all other values, including io.EOF come from the underlying reader.
//
func (u *Unframer) ReadUnescaped(p []byte) (n int, err error) {

	var buf []byte
	for len(p) > 0 {

		buf, err = u.r.Peek(len(p))
		switch err {
		case nil:
			break
		case io.EOF:
			if len(buf) == 0 {
				err = io.ErrUnexpectedEOF
			}
			fallthrough
		default:
			return
		}

		idx := find(buf, FLAG)
		idx_abrt := find(buf, ABORT)
		if idx_abrt < idx {
			idx = idx_abrt
		}

		if idx == 0 {
			break
		}

		// Make sure  we don't read a partial ESC.
		// this can happen when p[] is too small to contain the whole frame,
		// or when the input contains ESC FLAG which we never generate,
		// and which we'll decode to a single ESC at the end of the frame.
		if idx == len(p) && buf[idx-1] == ESC {
			p = p[:len(p)-1]
			idx--
		}

		var nn int
		nn, err = u.r.Read(p[:idx])
		m := unescape(p[:nn])
		if m > 0 {
			u.lastwasflag = false
		}
		p = p[m:]
		n += m
		if err != nil {
			return
		}
	}

	if len(p) == 0 {
		return n, ErrPartial
	}

	if len(buf) == 0 {
		panic("unescaping lost the flag")
	}

	if buf[0] == ABORT {
		_, err = u.r.Discard(1)
		if err != nil {
			// shouldn't happen; discarding already peeked byte
			panic("could not discard abort")
		}
		return n, ErrAbort
	}

	for len(buf) > 0 && buf[0] == FLAG {
		idx := findNot(buf, FLAG)
		var nn int
		nn, err = u.r.Discard(idx)
		if nn > 0 {
			u.lastwasflag = true
		}
		if err != nil {
			return
		}

		buf, err = u.r.Peek(bufSize)
		if err == io.EOF {
			continue
		}
		if err != nil {
			break
		}
	}
	return
}

// in-place unescaping of p.  returns its new length
// p should only end with ESC if it is the last byte in a frame
// in which case it is decoded as ESC
func unescape(p []byte) int {
	i, k := 0, 0
	for ; i < len(p)-1; i, k = i+1, k+1 {
		if p[i] == ESC {
			i++
			p[k] = p[i] ^ 0x20
		} else {
			p[k] = p[i]
		}
	}
	p[k] = p[i] // last one is copied, even if it is a raw ESC
	return k
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
