package dnsmsg

import "testing"

func TestDeserialize(t *testing.T) {
	b := []byte{0x06, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x00, 0x01, 0x00, 0x01, 0xc0, 0x0c, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x01, 0x2b, 0x00, 0x04, 0xd8, 0x3a, 0xd5, 0x6e}
	q := NewQuery("")
	q.Deserialize(b)
}
