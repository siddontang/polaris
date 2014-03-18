package session

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
)

type RedisStoreConfig struct {
	Password string `json:"password"`
	Addr     string `json:"addr"`
	DB       int    `json:"db"`

	//sesion max age, like http cookie max-age
	MaxAge int `json:"maxage"`

	//name registered for codec
	//if no name supplied, use default GobCodec for session encode/decode
	CodecName string `json:"codec"`

	//max idle connection for remote redis
	MaxIdle int `json:"maxidle"`
}

type RedisStore struct {
	config *RedisStoreConfig
	pool   *redis.Pool
	codec  Codec
}

func (store *RedisStore) Get(id string) (*Session, error) {
	var isNew bool = true
	var values = map[interface{}]interface{}{}

	if len(id) > 0 {
		c := store.pool.Get()

		buf, err := redis.Bytes(c.Do("GET", id))
		c.Close()

		if err != nil && err != redis.ErrNil {
			return nil, err
		}

		if buf != nil {
			values, err = store.codec.Decode(buf)
			if err != nil {
				return nil, err
			}

			isNew = false
		}
	}

	if isNew {
		id = GenerateID()
	}

	s := NewSession(id, store, store.config.MaxAge)

	s.Values = values

	return s, nil
}

func (store *RedisStore) Save(s *Session) error {
	buf, err := store.codec.Encode(s.Values)
	if err != nil {
		return err
	}
	c := store.pool.Get()
	_, err = c.Do("SETEX", s.ID(), s.MaxAge, buf)
	c.Close()

	return err
}

func (store *RedisStore) Delete(s *Session) error {
	c := store.pool.Get()
	_, err := c.Do("DEL", s.ID())
	c.Close()

	return err
}

type RedisDriver struct {
}

//json config: json encode RedisStoreConfig
func (d RedisDriver) Open(jsonConfig string) (Store, error) {
	cfg := new(RedisStoreConfig)

	if err := json.Unmarshal([]byte(jsonConfig), cfg); err != nil {
		return nil, err
	}

	f := func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", cfg.Addr)
		if err != nil {
			return nil, err
		}

		if len(cfg.Password) > 0 {
			if _, err = c.Do("AUTH", cfg.Password); err != nil {
				c.Close()
				return nil, err
			}
		}

		if cfg.DB != 0 {
			if _, err = c.Do("SELECT", cfg.DB); err != nil {
				c.Close()
				return nil, err
			}
		}

		return c, err
	}

	store := new(RedisStore)

	store.config = cfg
	store.codec = GetCodec(cfg.CodecName)

	store.pool = redis.NewPool(f, cfg.MaxIdle)

	return store, nil
}

func init() {
	Register("redis", RedisDriver{})
}
