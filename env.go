package polaris

import (
	"encoding/json"
	"net/http"
)

type Env struct {
	Request *http.Request
	Status  int

	w        http.ResponseWriter
	finished bool
}

func newEnv(w http.ResponseWriter, r *http.Request) *Env {
	e := new(Env)

	e.Request = r
	e.w = w

	e.finished = false
	e.Status = http.StatusOK

	return e
}

func (e *Env) Header() http.Header {
	return e.w.Header()
}

func (e *Env) SetContentType(tp string) {
	e.w.Header().Set("Content-type", tp)
}

func (e *Env) SetContentJson() {
	e.w.Header().Set("Content-type", "application/json; charset=utf-8")
}

func (e *Env) SetStatus(status int) {
	e.Status = status
}

func (e *Env) Write(v interface{}) {
	if e.finished {
		return
	}

	buf, err := json.Marshal(v)
	if err != nil {
		e.WriteError(http.StatusInternalServerError, err)
	} else {
		e.SetContentJson()

		e.WriteBuffer(buf)
	}
}

func (e *Env) WriteString(data string) {
	if e.finished {
		return
	}

	e.WriteBuffer([]byte(data))
}

func (e *Env) WriteBuffer(data []byte) {
	if e.finished {
		return
	}

	e.finished = true

	e.w.WriteHeader(e.Status)
	e.w.Write(data)
}

func (e *Env) WriteError(status int, err error) {
	e.Status = status
	e.WriteString(err.Error())
}

func (e *Env) finish() {
	if e.finished {
		return
	}

	e.finished = true

	e.w.WriteHeader(e.Status)
}
