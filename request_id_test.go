package bkit

import (
	"testing"
)

func TestNewRequestID(t *testing.T) {
	id := NewRequestID()
	idHex := id.Hex()
	time := id.Time()

	newID, err := new(RequestID).FromHex(idHex)
	if err != nil {
		t.Fatal(err)
	}
	if newID.Hex() != idHex {
		t.Fatal("newID.Hex() != idHex", newID.Hex(), idHex)
	}
	t.Logf("id: %s, time: %v", idHex, time)
}
