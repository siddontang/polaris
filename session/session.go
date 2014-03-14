package session

import (
	"fmt"
	"sync"
)

type Session interface {
	//session id
	ID() string

	//set value by key
	Set(key interface{}, value interface{}) error

	//get value by key, nil if key not exist
	Get(key interface{}) interface{}

	//delete value by key
	Delete(key interface{}) error

	//set session expire time, it will be affected after save
	Expire(seconds int) error

	//save session to store
	Save() error

	//delete all data and delete session from store
	Flush() error
}

type SessionStore interface {
	//get or new a session by id
	Get(id string) (Session, error)
}

var stores = map[string]func(*Config) (SessionStore, error){}
var storeLock sync.Mutex

func Register(storeName string, newStoreFn func(*Config) (SessionStore, error)) error {
	storeLock.Lock()
	defer storeLock.Unlock()

	if _, ok := stores[storeName]; ok {
		return fmt.Errorf("%s has been registered", storeName)
	}

	stores[storeName] = newStoreFn

	return nil
}

func Open(storeName string, cfg *Config) (SessionStore, error) {
	storeLock.Lock()
	defer storeLock.Unlock()

	fn, ok := stores[storeName]
	if !ok {
		return nil, fmt.Errorf("%s hasn't been registered", storeName)
	}

	return fn(cfg)
}
