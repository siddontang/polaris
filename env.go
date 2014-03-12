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

func (e *Env) SetStatus(status int) {
	e.Status = status
}

func (e *Env) Write(v interface{}) {
	if e.finished {
		return
	}

	e.finished = true

	buf, err := json.Marshal(v)
	if err != nil {
		http.Error(e.w, err.Error(), http.StatusInternalServerError)
		return
	}

	e.w.Header().Set("Content-type", "application/json; charset=utf-8")

	e.w.WriteHeader(e.Status)
	e.w.Write(buf)
}

func (e *Env) WriteError(status int, message string, result ...string) {
	var r string
	if len(result) == 0 || len(result[0]) == 0 {
		r = http.StatusText(status)
	} else {
		r = result[0]
	}

	e.Status = status
	e.Write(&HTTPError{status, message, r})
}

func (e *Env) finish() {
	if e.finished {
		return
	}

	e.finished = true

	e.w.WriteHeader(e.Status)
}
