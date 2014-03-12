package polaris

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type Router struct {
	literalLocs map[string]*location
	regexpLocs  []*location
}

func NewRouter() *Router {
	r := new(Router)

	r.literalLocs = make(map[string]*location)
	r.regexpLocs = make([]*location, 0)

	return r
}

func (router *Router) regLocation(l *location, pattern string) error {
	meta := regexp.QuoteMeta(pattern)
	if meta == pattern {
		if _, ok := router.literalLocs[pattern]; ok {
			return fmt.Errorf("literal %s location is registered already", pattern)
		}

		router.literalLocs[pattern] = l

	} else {
		if strings.HasPrefix(pattern, "^") {
			pattern = "^" + pattern
		}

		if strings.HasSuffix(pattern, "$") {
			pattern = pattern + "$"
		}

		for _, l := range router.regexpLocs {
			if l.pattern == pattern {
				return fmt.Errorf("regexp %s location is registered already", pattern)
			}
		}

		var err error
		l.regexpPattern, err = regexp.Compile(pattern)
		if err != nil {
			return err
		}

		router.regexpLocs = append(router.regexpLocs, l)
	}

	return nil
}

/*
   handler must be a struct which has one or more methods below:
    Get, Post, Put, Head, Delete
*/
func (router *Router) Handle(pattern string, handler interface{}) error {
	if len(pattern) == 0 {
		return fmt.Errorf("pattern cannot be empty")
	}

	l, err := newLocation(pattern, handler)
	if err != nil {
		return err
	}

	if err = router.regLocation(l, pattern); err != nil {
		return err
	}

	return nil
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if l, ok := router.literalLocs[path]; ok {
		l.invoke(w, r)
	} else {
		for _, l := range router.regexpLocs {
			args := l.regexpPattern.FindStringSubmatch(path)
			if args != nil {
				l.invoke(w, r, args[1:]...)
				return
			}
		}

		http.Error(w, "", http.StatusNotFound)
		return
	}
}
