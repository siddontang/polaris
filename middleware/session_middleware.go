package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/siddontang/polaris/context"
	"github.com/siddontang/polaris/session"
	"net/http"
)

const DefaultSecretKey = "polarissssiralop"

type SessionConfig struct {
	CookieName     string `json:"name"`
	CookiePath     string `json:"path"`
	CookieMaxAge   int    `json:"maxage"`
	CookieSecure   bool   `json:"secure"`
	CookieHttpOnly bool   `json:"httponly"`

	SecretKey string `json:"secret_key"`

	StoreName   string          `json:"store"`
	StoreConfig json.RawMessage `json:"store_config"`

	secretKey []byte
}

type SessionMiddleware struct {
	config *SessionConfig
	store  session.Store
}

func (m *SessionMiddleware) ProcessRequest(env *context.Env) error {
	if env.Session != nil {
		return fmt.Errorf("another session exist")
	}

	var id string
	if c, err := env.Request.Cookie(m.config.CookieName); err != nil {
		if err == http.ErrNoCookie {
			id = ""
		} else {
			return err
		}
	} else {
		id, err = DecodeSignID(c.Value, m.config.secretKey)
		if err != nil {
			return err
		}
	}

	var err error
	env.Session, err = m.store.Get(id)
	if err != nil {
		return err
	}

	return nil
}

func (m *SessionMiddleware) ProcessResponse(env *context.Env) error {
	if env.Session == nil {
		return fmt.Errorf("no session exist")
	}

	env.Session.Expire(m.config.CookieMaxAge)

	if err := env.Session.Save(); err != nil {
		return err
	}

	id, err := EncodeSignID(env.Session.ID(), m.config.secretKey)
	if err != nil {
		return err
	}

	c := &http.Cookie{
		Name:     m.config.CookieName,
		Value:    id,
		Path:     m.config.CookiePath,
		MaxAge:   m.config.CookieMaxAge,
		Secure:   m.config.CookieSecure,
		HttpOnly: m.config.CookieHttpOnly,
	}
	env.SetCookie(c)

	return nil
}

type SessoionMiddlewareDriver struct {
}

func (d SessoionMiddlewareDriver) Open(jsonConfig json.RawMessage) (Middleware, error) {
	config := new(SessionConfig)

	if err := json.Unmarshal(jsonConfig, config); err != nil {
		return nil, err
	}

	if len(config.SecretKey) == 0 {
		config.SecretKey = DefaultSecretKey
	} else if len(config.SecretKey)%16 != 0 {
		return nil, fmt.Errorf("invalid secret key len %d, must multi 16", len(config.SecretKey))
	}

	config.secretKey = []byte(config.SecretKey)

	m := new(SessionMiddleware)

	var err error
	m.store, err = session.Open(config.StoreName, config.StoreConfig)
	if err != nil {
		return nil, err
	}

	m.config = config

	return m, nil
}

func init() {
	Register("session", SessoionMiddlewareDriver{})
}
