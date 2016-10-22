#ifndef __HDLC__HDLC__H_
#define __HDLC__HDLC__H_

#include "stddef.h"
#include "stdint.h"

enum {
	HDLC_ESC   = 0x7d,
	HDLC_FLAG  = 0x7e,
	HDLC_ABORT = 0x7f,
};

/** hdlc_encode returns the next byte to transmit the HDLC framing of the given buffer.
 *
 * Note: this modifies the passed in buffer! (bit 5 of the escaped values is flipped.)
 *
 * Start a transmission by sending HDLC_FLAG (0x7e) and call
 *
 *    b = hdlc_encode(&bufp, buf_end)
 *
 * to obtain the next byte b to send until b == HDLC_FLAG again.
 *
 * The closing flag of a frame may be used as the starting flag of the next frame.
 * The flag may also be used as a fill value between frames.
 *
 * When *bufp == buf_end, hdlc_encode returns HDLC_FLAG, otherwise
 * when **bufp is in {0x7d, 0x7e, 0x7f}, it returns HDLC_ESC(0x7d) after modifying
 * **bufp to 0x5{d,e,f}. Otherwise hdlc_encode returns **bufp and moves *bufp one
 * ahead.
 *
 * To abort a partially sent frame, just send HDLC_ABORT(0x7f).
 *
 */
uint8_t hdlc_encode(uint8_t **bufpp, uint8_t *buf_end);

/** hdlc_decode appends the received unframed value val to the provided buffer.
 *
 *  The receive buffer must be initialized with a 0 value at **bufpp.
 *
 *  If *bufpp < buf_end and the received value is HDLC_ESC
 *
 * Return values
 *
 *    0           the the frame is not yet complete
 *    HDLC_FLAG   the received frame is complete
 *    HDLC_ABORT  the received frame was aborted by the sender
 *                OR the buffer is full.
 *  Examine if *bufpp == buf_end to distinguish between the two HLDC_ABORT causes.
 *
 * to re-sync to a frame start, just wait until you receive HDLC_FLAG (0x7e) before
 * starting to call hdlc_decode. the first call to hdlc_decode should be after
 *
 * example:
 *            int recv;
 *        resync:
 *            while ((recv = get_char()) != EOF)
 *                if (recv == HDLC_FLAG)
 *                    break;
 *
 *        	  if (recv == EOF)
 *                 return;
 *
 *            int flag;
 *        packet:
 *            while ((flag = hdlc_decode(&bufp, bufe, recv)) == 0)
 *                if ((recv = get_char()) == EOF)
 *                    return;
 *
 *       // distinguish:
 *       //  flag == HDLC_FLAG  -> goto packet
 *       //  flag == HDLC_ABORT -> goto resync
 *
 *
 */
int hdlc_decode(uint8_t **bufpp, uint8_t *buf_end, uint8_t val);

#endif //__HDLC__HDLC__H_
