// Package hdlc implements HDLC-like framing of packets on bytestreams.
//
// The go io.Writer and io.Reader interfaces deal with turning slices into byte streams.
// The semantics of the streams are up to the implementations, but almost always the
// structure of the slices that were passed is lost, e.g. when you write two slices
// to an os.File and then read back the file the information on the separation between
// the two write calls is lost, (as it should).
//
// Putting the size or boundary information in is called framing. A standard way is to prefix
// some length information in-band and read that back or terminate or separate packets
// with a reserved value.  Occurences of this value in the payload have to be disambiguated
// from the terminator in that case.
//
// HDLC(-like) framing works as follows:
//
// In the stream of bytes, ESC, FLAG and ABORT (0x7d, 0x7e and 0x7f) are reserved.
// Whenever they occur in the packet they are replaced by the 2-byte sequence ESC-(val ^ 0x20).
// This frees up the FLAG and ABORT character: FLAG is used to separate frames, and ABORT is
// used to terminate a partial transmission.
//
// This works well for best-effort communication over serial byte streams, since it
// is always easy to resync to a frame boundary.
//
// This package does not deal with the address, control and crc fields of a true HDLC frame.
// You can add those by binary-encoding them into the buffers you send.
//
package hdlc

// ESC, FLAG and ABORT are escaped in the stream by prefixing ESC and flipping bit 5.
const (
	ESC   = 0x7d
	FLAG  = 0x7e // Frame separator
	ABORT = 0x7f // Abort a frame. Receiver should discard.
)
