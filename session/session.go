package session

import (
	"fmt"
	"sync"
)

type Session interface {
	ID() string

	Set(key interface{}, value interface{}) error
	Get(key interface{}) interface{}
	Delete(key interface{}) error

	Save() error
	Flush() error
}

type SessionStore interface {
	Get(id string) (Session, error)
}

var stores map[string]func() (SessionStore, error)
var storeLock sync.Mutex

func Register(storeName string, newStoreFn func() (SessionStore, error)) error {
	storeLock.Lock()
	defer storeLock.Unlock()

	if _, ok := stores[storeName]; ok {
		return fmt.Errorf("%s has been registered", storeName)
	}

	stores[storeName] = newStoreFn

	return nil
}

func Open(storeName string) (SessionStore, error) {
	storeLock.Lock()
	defer storeLock.Unlock()

	fn, ok := stores[storeName]
	if !ok {
		return nil, fmt.Errorf("%s hasn't been registered", storeName)
	}

	return fn()
}
