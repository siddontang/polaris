package session

import (
	"fmt"
)

type Store interface {
	//get a session by id
	//if no session exist, regenerate another id to new a session
	Get(id string) (*Session, error)

	//delete session from store
	Delete(*Session) error

	//Save session to stroe
	Save(*Session) error
}

type Driver interface {
	Open(dsn string) (Store, error)
}

var stores = map[string]Driver{}

func Register(name string, d Driver) error {
	if _, ok := stores[name]; ok {
		return fmt.Errorf("%s has been registered", name)
	}

	stores[name] = d
	return nil
}

func Open(name string, config string) (Store, error) {
	if f, ok := stores[name]; ok {
		return f.Open(config)
	} else {
		return nil, fmt.Errorf("%s has not been registered", name)
	}
}
