package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/siddontang/polaris/context"
)

type Middleware interface {
	ProcessRequest(env *context.Env) error
	ProcessResponse(env *context.Env) error
}

type MiddlewareDriver interface {
	Open(jsonConfig json.RawMessage) (Middleware, error)
}

var middles = map[string]MiddlewareDriver{}

func Register(name string, driver MiddlewareDriver) error {
	if _, ok := middles[name]; ok {
		return fmt.Errorf("middleware %s has been registered", name)
	}

	middles[name] = driver
	return nil
}

func Open(name string, configJson json.RawMessage) (Middleware, error) {
	if m, ok := middles[name]; ok {
		return m.Open(configJson)
	} else {
		return nil, fmt.Errorf("middleware %s has not been registered", name)
	}
}
