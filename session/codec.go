package session

import (
	"bytes"
	"encoding/gob"
)

type Codec interface {
	Encode(values map[interface{}]interface{}) ([]byte, error)
	Decode(buf []byte) (map[interface{}]interface{}, error)
}

type GobCodec struct {
}

func (c GobCodec) Encode(values map[interface{}]interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(values)
	if err != nil {
		return []byte(""), err
	}
	return buf.Bytes(), nil
}

func (c GobCodec) Decode(data []byte) (map[interface{}]interface{}, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	var out map[interface{}]interface{}
	err := dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func init() {
	gob.Register(map[interface{}]interface{}{})
}
