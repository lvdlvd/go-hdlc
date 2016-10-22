// gcc hdlc.c hdlc_test.c && ./a.out

#include "hdlc.h"
#include <stdio.h>
#include <stdlib.h>

int main(int argc, char *argv[]) {

	uint8_t bufxmit[800];
	uint8_t bufrecv[801]; // need 1 extra for closing flag

	for (int i = 0; i < 800; ++i) {
		bufxmit[i] = lrand48();
	}

	bufrecv[0] = 0;

	uint8_t *txp = bufxmit;
	uint8_t *rxp = bufrecv;

	for (;;) {
		uint8_t val  = hdlc_encode(&txp, bufxmit + 800);
		int     flag = hdlc_decode(&rxp, bufrecv + 801, val);
		if (flag != 0) {
			break;
		}
	}

	int err = 0;
	for (int i = 0; i < 800; ++i) {
		// unflip escaped values
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