package session

import ()

type Session struct {
	Values map[interface{}]interface{}

	MaxAge int

	id    string
	store Store
}

//session id
func (s *Session) ID() string {
	return s.id
}

//set value by key
func (s *Session) Set(key interface{}, value interface{}) error {
	s.Values[key] = value
	return nil
}

//get value by key, nil if key not exist
func (s *Session) Get(key interface{}) interface{} {
	if v, ok := s.Values[key]; ok {
		return v
	} else {
		return nil
	}
}

//delete value by key
func (s *Session) Delete(key interface{}) error {
	delete(s.Values, key)
	return nil
}

//set session expire time, it will be affected after save
func (s *Session) Expire(seconds int) error {
	if seconds < 0 {
		seconds = 0
	}

	s.MaxAge = seconds
	return nil
}

//save session to store
func (s *Session) Save() error {
	return s.store.Save(s)
}

//delete all data, delete session from store and regenerate id
func (s *Session) Flush() error {
	if err := s.store.Delete(s); err != nil {
		return err
	}

	s.Values = make(map[interface{}]interface{})

	s.id = GenerateID()
	return nil
}

func NewSession(id string, store Store, maxAge int) *Session {
	s := new(Session)

	s.id = id
	s.store = store
	s.MaxAge = maxAge

	s.Values = make(map[interface{}]interface{})

	return s
}
