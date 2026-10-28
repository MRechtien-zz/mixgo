package nrpn

import (
	"bytes"
	"testing"
)

func TestMessage(t *testing.T) {
	got := Message(3, 0x04, 0x02, 0x01, 0x7F)
	want := []byte{
		0xB3, 0x63, 0x04,
		0xB3, 0x62, 0x02,
		0xB3, 0x06, 0x01,
		0xB3, 0x26, 0x7F,
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("Message = % X, want % X", got, want)
	}
}

func TestMessageMasksChannel(t *testing.T) {
	got := Message(0xFF, 0, 0, 0, 0)
	if got[0] != 0xBF {
		t.Fatalf("status = %X, want BF", got[0])
	}
}
