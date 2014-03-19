package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"fmt"
)

func EncodeSignID(id string, key []byte) (string, error) {
	if len(id) == 0 {
		return "", fmt.Errorf("empty id not allowed")
	}

	s := hmac.New(md5.New, key)
	s.Write([]byte(id))

	buf := append(s.Sum(nil), id...)

	return base64.StdEncoding.EncodeToString(buf), nil
}

func DecodeSignID(src string, key []byte) (string, error) {
	buf, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return "", err
	}

	if len(buf) <= 16 {
		return "", fmt.Errorf("invalid sign id length")
	}

	s := hmac.New(md5.New, key)
	s.Write(buf[16:])

	if !bytes.Equal(s.Sum(nil), buf[:16]) {
		return "", fmt.Errorf("invalid sign id")
	}

	return string(buf[16:]), nil
}
