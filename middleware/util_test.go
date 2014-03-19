package middleware

import (
	"testing"
)

func TestSignID(t *testing.T) {
	id := "abcdefghabcdefgh"

	key := []byte("1234567887654321")

	s, err := EncodeSignID(id, key)
	if err != nil {
		t.Fatal(err)
	}

	var id2 string
	id2, err = DecodeSignID(s, key)
	if err != nil {
		t.Fatal(err)
	}

	if id != id2 {
		t.Fatal(id, id2)
	}
}
