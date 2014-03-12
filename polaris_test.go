package polaris

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

type TestHandler1 struct {
}

func (h *TestHandler1) Prepare(env *Env) {

}

func (h *TestHandler1) Get(env *Env) {

}

type TestHandler2 struct {
}

func (h *TestHandler2) Get(env *Env, id string) {
	v := struct {
		ID   string
		Name string
	}{
		id,
		"hello",
	}

	env.Write(v)
}

type TestHandler3 struct {
}

func (h *TestHandler3) Get(env *Env) {
	env.WriteError(http.StatusForbidden, "forbidden")
}

func TestPolaris(t *testing.T) {
	r := NewRouter()

	if err := r.Handle("/test1", new(TestHandler1)); err != nil {
		t.Fatal(err)
	}

	if err := r.Handle("/test2/([0-9]+)", new(TestHandler2)); err != nil {
		t.Fatal(err)
	}

	if err := r.Handle("/test3", new(TestHandler3)); err != nil {
		t.Fatal(err)
	}

	http.Handle("/", r)
	go http.ListenAndServe("127.0.0.1:11181", nil)

	time.Sleep(1 * time.Second)

	var test1 interface{}
	if err := testRequest("GET", "/test1", nil, &test1); err != nil {
		t.Fatal(err)
	} else if test1 != nil {
		t.Fatal(test1)
	}

	var test2 struct {
		ID   string
		Name string
	}

	if err := testRequest("GET", "/test2/1234", nil, &test2); err != nil {
		t.Fatal(err)
	} else {
		if test2.ID != "1234" {
			t.Fatal(test2.ID)
		}

		if test2.Name != "hello" {
			t.Fatal(test2.Name)
		}
	}

	var test3 interface{}
	if err := testRequest("GET", "/test3", nil, &test3); err == nil {
		t.Fatal("must error")
	} else {
		if e, ok := err.(HTTPError); !ok {
			t.Fatal("must http error")
		} else {
			if e.Status != http.StatusForbidden {
				t.Fatal(e.Status)
			}

			if e.Message != "forbidden" {
				t.Fatal(e.Message)
			}
		}
	}
}

func testRequest(method string, path string, data io.Reader, v interface{}) error {
	url := "http://127.0.0.1:11181" + path
	r, err := http.NewRequest(method, url, data)
	if err != nil {
		return err
	}

	var resp *http.Response
	client := http.Client{}
	resp, err = client.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)

	var e HTTPError
	if resp.StatusCode != http.StatusOK {
		e.Status = resp.StatusCode

		err = json.Unmarshal(body, &e)
		if err != nil {
			e.Message = string(body)
		}
		return e
	} else {
		if len(body) == 0 {
			v = nil
			return nil
		}

		err = json.Unmarshal(body, v)
		if err != nil {
			return err
		}
	}

	return nil
}
