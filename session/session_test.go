package session

import (
	"testing"
)

var testConfig = &Config{
	Timeout:    3600,
	Serializer: GobCodec{},
	Redis: RedisConfig{
		DSN:     "10.20.187.120:6379",
		MaxIdle: 16,
	},
}

func TestRedisSession(t *testing.T) {
	Register("redis", NewRedisSessionStore)

	if err := Register("redis", NewRedisSessionStore); err == nil {
		t.Fatal("must error")
	}

	store, err := Open("redis", testConfig)
	if err != nil {
		t.Fatal(err)
	}

	if session, err := store.Get("polaris:session:1"); err != nil {
		t.Fatal(err)
	} else {
		session.Set("1", "Hello")

		if err := session.Save(); err != nil {
			t.Fatal(err)
		}
	}

	if session, err := store.Get("polaris:session:1"); err != nil {
		t.Fatal(err)
	} else {
		if value := session.Get("1"); value != "Hello" {
			t.Fatal(value)
		}

		session.Flush()
	}

	if session, err := store.Get("polaris:session:1"); err != nil {
		t.Fatal(err)
	} else {
		if value := session.Get("1"); value != nil {
			t.Fatal(value)
		}
	}
}
