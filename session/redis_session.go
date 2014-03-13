package session

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"strings"
	"sync"
)

type RedisSession struct {
	sync.Mutex
	id   string
	data map[interface{}]interface{}

	store *RedisSessionStore
}

func (s *RedisSession) ID() string {
	return s.id
}

func (s *RedisSession) Set(key interface{}, value interface{}) error {
	s.Lock()
	s.data[key] = value
	s.Unlock()
	return nil
}

func (s *RedisSession) Get(key interface{}) interface{} {
	s.Lock()
	v, ok := s.data[key]
	s.Unlock()
	if ok {
		return v
	} else {
		return nil
	}
}

func (s *RedisSession) Delete(key interface{}) error {
	s.Lock()
	delete(s.data, key)
	s.Unlock()
	return nil
}

func (s *RedisSession) saveTimeout(c redis.Conn, buf []byte) error {
	if err := c.Send("SET", s.id, buf); err != nil {
		return err
	}

	if err := c.Send("EXPIRE", s.id, s.store.timeout); err != nil {
		return err
	}

	if err := c.Flush(); err != nil {
		return err
	}

	if _, err := c.Receive(); err != nil {
		return err
	}

	if _, err := c.Receive(); err != nil {
		return err
	}

	return nil
}

func (s *RedisSession) Save() error {
	s.Lock()

	buf, err := json.Marshal(s.data)
	if err != nil {
		s.Unlock()
		return err
	}

	s.Unlock()

	c := s.store.pool.Get()

	if s.store.timeout <= 0 {
		_, err = c.Do("SET", s.id, buf)
	} else {
		err = s.saveTimeout(c, buf)
	}
	c.Close()

	return err
}

func (s *RedisSession) Flush() error {
	s.Lock()

	s.data = make(map[interface{}]interface{})

	s.Unlock()

	c := s.store.pool.Get()
	_, err := c.Do("DEL", s.id)
	c.Close()

	return err
}

type RedisSessionStore struct {
	pool    *redis.Pool
	timeout int
}

func (store *RedisSessionStore) Get(id string) (Session, error) {
	c := store.pool.Get()

	buf, err := redis.Bytes(c.Do("GET", id))
	c.Close()

	if err != nil && err != redis.ErrNil {
		return nil, err
	}

	s := new(RedisSession)
	s.store = store
	s.id = id
	s.data = make(map[interface{}]interface{})

	if buf == nil {
		return s, nil
	}

	err = json.Unmarshal(buf, &s.data)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func parseRedisDSN(dsn string) (password string, addr string, db int, err error) {
	seps := strings.Split(dsn, "@")
	if len(seps) > 2 || len(seps) == 0 {
		err = fmt.Errorf("invalid dsn %s, need:password>@<host>:<port>/<db>", dsn)
		return
	}

	var left string = seps[0]
	if len(seps) == 2 {
		password = seps[0]
		left = seps[1]
	}

	seps = strings.Split(left, "/")
	if len(seps) > 2 || len(seps) == 0 {
		err = fmt.Errorf("invalid dsn %s, need:password>@<host>:<port>/<db>", dsn)
		return
	}

	db = 0
	if len(seps) == 2 {
		if db, err = strconv.Atoi(seps[1]); err != nil {
			return
		}
	}
	return
}

/*
   redis data source name(dsn):
   <password>@<host>:<port>/<db>
*/
func NewRedisSessionStore(dsn string, maxIdle int, timeout int) (SessionStore, error) {
	password, addr, db, err := parseRedisDSN(dsn)
	if err != nil {
		return nil, err
	}

	f := func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}

		if len(password) > 0 {
			if _, err = c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}
		}

		if db != 0 {
			if _, err = c.Do("SELECT", db); err != nil {
				c.Close()
				return nil, err
			}
		}

		return c, err
	}

	store := new(RedisSessionStore)

	store.timeout = timeout

	store.pool = redis.NewPool(f, maxIdle)

	return store, nil
}
