// gcc hdlc.c hdlc_test.c && ./a.out

#include "hdlc.h"
#include <stdio.h>
#include <stdlib.h>

int main(int argc, char *argv[]) {

	uint8_t bufxmit[800];
	uint8_t bufrecv[801]; // need 1 extra for closing flag

	// Fill bufxmit with random garbage
	for (int i = 0; i < 800; ++i) {
		bufxmit[i] = lrand48();
	}

	// hdlc_decode keeps state in the next character, so clear that
	bufrecv[0] = 0;

	uint8_t *txp = bufxmit;
	uint8_t *rxp = bufrecv;

	// transmit the message from bufxmit to bufrecv
	for (;;) {
		uint8_t val  = hdlc_encode(&txp, bufxmit + 800);
		int     flag = hdlc_decode(&rxp, bufrecv + 801, val);
		if (flag != 0) {
			break;
		}
	}
	
	int err = 0;
	for (int i = 0; i < 800; ++i) {
		// hdlc_encode will have flipped bit 5 of escaped values in bufxmit
		// so unflip escaped values again before comparing
		switch (bufrecv[i]) {
		case HDLC_ESC:
		case HDLC_FLAG:
		case HDLC_ABORT:
			bufxmit[i] ^= 0x20;
		}

		if (bufxmit[i] != bufrecv[i]) {
			printf("difference at %d: %x %x\n", i, bufxmit[i], bufrecv[i]);
			++err;
		}
	}

	if (err == 0) {
		printf("ok.");
	}

	return 0;
}
