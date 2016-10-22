#include "hdlc.h"

uint8_t hdlc_encode(uint8_t **bufpp, uint8_t *buf_end) {
	if (*bufpp == buf_end) {
		return HDLC_FLAG;
	}

	uint8_t c = **bufpp;
	switch (c) {
	case HDLC_ESC:
	case HDLC_FLAG:
	case HDLC_ABORT:
		**bufpp = c ^ 0x20;
		return HDLC_ESC;
	}

	++(*bufpp);
	return c;
}

int hdlc_decode(uint8_t **bufpp, uint8_t *buf_end, uint8_t val) {

	if (*bufpp == buf_end) {
		return HDLC_ABORT;
	}

	uint8_t c = **bufpp; // current state
	switch (c) {
	case HDLC_ESC:
		// previous val was an ESC
		**bufpp = val ^ 0x20;
		break;

	case 0:                     // initial value at buffer begin
		if (val == HDLC_FLAG) { // repeated FLAG, no change
			return 0;
		}
	// fallthrough

	case HDLC_FLAG:
		// previous val was something unescaped, or at the start of the buffer
		**bufpp = val;

		switch (val) {
		case HDLC_ESC:
			// wrote the esc state, don't change pointer
			return 0;

		case HDLC_FLAG:
		case HDLC_ABORT:
			return val;
		}
		break;

	default:
		// corrupt buffer
		return HDLC_ABORT;
	}

	// if we're here, we wrote a value and we're not done with the frame
	++(*bufpp);

	// no room for state
	if (*bufpp == buf_end) {
		return HDLC_ABORT;
	}

	**bufpp = HDLC_FLAG; // prepare state for next received value
	return 0;
}
