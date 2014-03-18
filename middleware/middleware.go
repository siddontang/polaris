package middleware

import (
	"fmt"
	"github.com/siddontang/polaris/context"
)

type Middleware interface {
	ProcessRequest(env *context.Env) error
	ProcessResponse(env *context.Env) error
}

type MiddlewareDriver interface {
	Open(config string) (Middleware, error)
}

var middles = map[string]MiddlewareDriver{}

func Register(name string, driver MiddlewareDriver) error {
	if _, ok := middles[name]; ok {
		return fmt.Errorf("middleware %s has been registered", name)
	}

	middles[name] = driver
	return nil
}

func Open(name string, config string) (Middleware, error) {
	if m, ok := middles[name]; ok {
		return m.Open(config)
	} else {
		return nil, fmt.Errorf("middleware %s has not been registered", name)
	}
}
