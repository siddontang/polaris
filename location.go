package polaris

import (
	"fmt"
	"github.com/siddontang/polaris/context"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

type location struct {
	pattern       string
	regexpPattern *regexp.Regexp
	methods       map[string]reflect.Value
	app           *App
}

func (l *location) invoke(w http.ResponseWriter, r *http.Request, args ...string) {
	env := context.NewEnv(w, r)

	defer func() {
		if err := recover(); err != nil {
			const size = 4096
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]

			env.WriteError(http.StatusInternalServerError,
				fmt.Errorf("panic: %v\n %v", err, buf))
		}

		env.Finish()
	}()

	if err := l.app.processRequest(env); err != nil {
		env.WriteError(http.StatusInternalServerError, err)
		return
	}

	envValue := reflect.ValueOf(env)

	m, ok := l.methods[r.Method]
	if !ok {
		env.SetStatus(http.StatusMethodNotAllowed)
		return
	}

	if prepare, ok := l.methods["PREPARE"]; ok {
		prepare.Call([]reflect.Value{envValue})
		if env.IsFinished() {
			return
		}
	}

	inNum := m.Type().NumIn()
	if inNum != len(args)+1 {
		env.WriteError(http.StatusForbidden, fmt.Errorf("%s input arguments %d != %d", r.Method, inNum-1, len(args)))
		return
	}

	in := make([]reflect.Value, m.Type().NumIn())

	in[0] = envValue

	for i, v := range args {
		in[i+1] = reflect.ValueOf(v)
	}

	m.Call(in)

	if err := l.app.processResponse(env); err != nil {
		env.WriteError(http.StatusInternalServerError, err)
		return
	}
}

var SupportMethods = []string{"Prepare", "Get", "Post", "Put", "Head", "Delete"}

//method first input argument must *http.Request
//method last output argument must be error interface{}
func (l *location) checkMethod(handler interface{}, m reflect.Type, name string) error {
	nIn := m.NumIn()

	if nIn == 0 || m.In(0).Kind() != reflect.Ptr {
		return fmt.Errorf("%T:function %s first input argument must *context.Env", handler, name)
	}

	if m.In(0).String() != "*context.Env" {
		return fmt.Errorf("%T:function %s first input argument must *context.Env", handler, name)
	}

	if name == "Prepare" && nIn > 1 {
		return fmt.Errorf("%T:function %s must have one input argument", handler, name)
	}

	for i := 1; i < nIn; i++ {
		//left arguments must be string
		if m.In(i).Kind() != reflect.String {
			return fmt.Errorf("%T:function %s %d input arguments must be string", handler, name, i)
		}
	}

	return nil
}

func newLocation(pattern string, handler interface{}) (*location, error) {
	v := reflect.ValueOf(handler)

	l := new(location)

	l.methods = make(map[string]reflect.Value)
	l.pattern = pattern

	var hasMethod bool = false
	for _, n := range SupportMethods {
		m := v.MethodByName(n)
		if m.Kind() == reflect.Func {
			if err := l.checkMethod(handler, m.Type(), n); err != nil {
				return nil, err
			}

			hasMethod = true

			l.methods[strings.ToUpper(n)] = m
		}
	}

	if !hasMethod {
		return nil, fmt.Errorf("handler has no [Get, Post, Put, Head, Delete] methods")
	}

	return l, nil
}
