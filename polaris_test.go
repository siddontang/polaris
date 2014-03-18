package polaris

import (
	"encoding/json"
	"github.com/siddontang/polaris/context"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

type TestHandler1 struct {
}

func (h *TestHandler1) Prepare(env *context.Env) {

}

func (h *TestHandler1) Get(env *context.Env) {

}

type TestHandler2 struct {
}

func (h *TestHandler2) Get(env *context.Env, id string) {
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

type TestHTTPError struct {
	Status  int    `json:"-"`
	Message string `json:"msg:`
	Result  string `json:"result"`
}

func (e TestHTTPError) Error() string {
	buf, _ := json.Marshal(e)
	return string(buf)
}

func (h *TestHandler3) Get(env *context.Env) {
	env.WriteError(http.StatusForbidden, TestHTTPError{http.StatusForbidden, "forbidden", "error"})
}

func TestPolaris(t *testing.T) {
	app, err := NewApp("./etc/polaris.json")
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Handle("/test1", new(TestHandler1)); err != nil {
		t.Fatal(err)
	}

	if err := app.Handle("/test2/([0-9]+)", new(TestHandler2)); err != nil {
		t.Fatal(err)
	}

	if err := app.Handle("/test3", new(TestHandler3)); err != nil {
		t.Fatal(err)
	}

	go app.Run()

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
		if e, ok := err.(TestHTTPError); !ok {
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

	var e TestHTTPError
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
