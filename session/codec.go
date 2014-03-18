package session

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"
)

//codec for session encode and decode
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

var codecs = map[string]Codec{}
var codecLock sync.Mutex

func RegisterCodec(name string, codec Codec) error {
	codecLock.Lock()
	defer codecLock.Unlock()

	if _, ok := codecs[name]; ok {
		return fmt.Errorf("%s has been registered", name)
	}

	codecs[name] = codec
	return nil
}

//get codec by name, or default codec(GobCodec) if name not exist
func GetCodec(name string) Codec {
	if c, ok := codecs[name]; ok {
		return c
	} else {
		return GobCodec{}
	}
}

func init() {
	gob.Register(map[interface{}]interface{}{})

	RegisterCodec("gob", GobCodec{})
}
