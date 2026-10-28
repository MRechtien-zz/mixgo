// Package nrpn builds MIDI NRPN message sequences.
package nrpn

// Message returns a 12 byte NRPN sequence:
//
//	CC 0x63 msb, CC 0x62 lsb, CC 0x06 coarse, CC 0x26 fine
//
// channel is zero based (0-15).
func Message(channel, msb, lsb, coarse, fine uint8) []byte {
	status := 0xB0 | (channel & 0x0F)
	return []byte{
		status, 0x63, msb,
		status, 0x62, lsb,
		status, 0x06, coarse,
		status, 0x26, fine,
	}
}
