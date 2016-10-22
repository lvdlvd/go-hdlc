# go-hdlc provides an io.Writer and an io.Reader that frame and unframe buffers.

The go io.Writer and Reader interfaces deal with turning slices into byte streams.
The semantics of the streams are up to the implementations, but almost always the 
structure of the slices that were passed is lost, eg when you write two slices
to an os.File and then read back the file the information on the separation between
the two calls is lost, (as it should).

Putting the block size information back in is called framing.  A standard way
is to output some length information in-band and read that back but for
some applications, eg communication over noisy serial lines, a technique called
HDLC(-like) framing is more appropriate. It works as follows:

In the stream of bytes, some values, in our case ESC, FLAG and ABORT (0x7d, 0x7e 
and 0x7f) are reserved.  Whenever they occur in the buffer they are replaced by the
2-byte sequence ESC-(val ^ 0x20).  This frees up the FLAG and ABORT character
since they are now guaranteed not to occur in the framed bytestream.  FLAG is used
to separate frames, and ABORT is used to terminate one half way.

the C directory contains 2 C functions to do the same.

I use this to have my go programs on my macbook talk to my arduino over serial ports.

