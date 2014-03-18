package session

import (
	"encoding/json"
	"testing"
)

func TestRedisSession(t *testing.T) {
	var config = RedisStoreConfig{
		Password:  "",
		Addr:      "127.0.0.1:6379",
		DB:        0,
		MaxAge:    3600,
		CodecName: "gob",
		MaxIdle:   16,
	}

	s, _ := json.Marshal(&config)

	store, err := Open("redis", string(s))
	if err != nil {
		t.Fatal(err)
	}

	var id string
	if session, err := store.Get("unkown session key"); err != nil {
		t.Fatal(err)
	} else {
		session.Set("1", "Hello")

		if err := session.Save(); err != nil {
			t.Fatal(err)
		}

		id = session.ID()
	}

	if session, err := store.Get(id); err != nil {
		t.Fatal(err)
	} else {
		if value := session.Get("1"); value != "Hello" {
			t.Fatal(value)
		}

		session.Flush()
	}

	if session, err := store.Get(id); err != nil {
		t.Fatal(err)
	} else {
		if value := session.Get("1"); value != nil {
			t.Fatal(value)
		}
	}
}
