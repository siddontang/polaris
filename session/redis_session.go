package session

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"strings"
	"sync"
)

type RedisSession struct {
	sync.Mutex
	id      string
	values  map[interface{}]interface{}
	dirty   bool
	timeout int

	store *RedisSessionStore
}

func (s *RedisSession) ID() string {
	return s.id
}

func (s *RedisSession) Set(key interface{}, value interface{}) error {
	s.Lock()
	s.dirty = true
	s.values[key] = value
	s.Unlock()
	return nil
}

func (s *RedisSession) Get(key interface{}) interface{} {
	s.Lock()
	v, ok := s.values[key]
	s.Unlock()
	if ok {
		return v
	} else {
		return nil
	}
}

func (s *RedisSession) Delete(key interface{}) error {
	s.Lock()
	s.dirty = true
	delete(s.values, key)
	s.Unlock()
	return nil
}

func (s *RedisSession) Expire(seconds int) error {
	s.Lock()
	s.dirty = true
	s.timeout = seconds
	s.Unlock()
	return nil
}

func (s *RedisSession) Save() error {
	s.Lock()
	if !s.dirty {
		s.Unlock()
		return nil
	}

	buf, err := s.store.codec.Encode(s.values)
	if err != nil {
		s.Unlock()
		return err
	}

	s.dirty = false

	s.Unlock()

	c := s.store.pool.Get()

	if s.timeout <= 0 {
		_, err = c.Do("SET", s.id, buf)
	} else {
		_, err = c.Do("SETEX", s.id, s.timeout, buf)
	}
	c.Close()

	return err
}

func (s *RedisSession) Flush() error {
	s.Lock()

	s.values = make(map[interface{}]interface{})
	s.dirty = true

	s.Unlock()

	c := s.store.pool.Get()
	_, err := c.Do("DEL", s.id)
	c.Close()

	return err
}

type RedisSessionStore struct {
	pool    *redis.Pool
	timeout int
	codec   Codec
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
	s.dirty = false
	s.values = make(map[interface{}]interface{})
	s.timeout = store.timeout

	if buf == nil {
		return s, nil
	}

	s.values, err = store.codec.Decode(buf)
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

	addr = seps[0]

	db = 0
	if len(seps) == 2 {
		if db, err = strconv.Atoi(seps[1]); err != nil {
			return
		}
	}
	return
}

func NewRedisSessionStore(cfg *Config) (SessionStore, error) {
	timeout := cfg.Timeout
	codec := cfg.Serializer
	dsn := cfg.Redis.DSN
	maxIdle := cfg.Redis.MaxIdle

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
	if codec == nil {
		store.codec = GobCodec{}
	} else {
		store.codec = codec
	}

	store.pool = redis.NewPool(f, maxIdle)

	return store, nil
}

func init() {
	Register("redis", NewRedisSessionStore)
}
