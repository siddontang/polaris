package polaris

import (
	"errors"
	"github.com/siddontang/polaris/context"
	"github.com/siddontang/polaris/middleware"
	"net/http"
)

var (
	ErrAppRunning = errors.New("app is running, cannot set")
)

type App struct {
	config      *Config
	running     bool
	router      *Router
	middlewares []middleware.Middleware
}

func NewApp(configFile string) (*App, error) {
	c, err := ParserConfig(configFile)
	if err != nil {
		return nil, err
	}

	app := new(App)
	app.running = false

	app.config = c
	app.router = newRouter(app)
	app.middlewares = make([]middleware.Middleware, len(c.Middlewares))

	for i, _ := range app.middlewares {
		app.middlewares[i], err = middleware.Open(c.Middlewares[i].Name,
			string(c.Middlewares[i].Config))
		if err != nil {
			return nil, err
		}
	}

	return app, nil
}

func (app *App) Config() *Config {
	return app.config
}

func (app *App) Handle(pattern string, handler interface{}) error {
	if app.running {
		return ErrAppRunning
	}

	return app.router.Handle(pattern, handler)
}

func (app *App) Run() error {
	app.running = true

	go func() {
		http.Handle("/", app.router)

		http.ListenAndServe(app.config.HttpAddr, nil)

	}()

	return nil
}

func (app *App) processRequest(env *context.Env) error {
	var err error
	for _, m := range app.middlewares {
		if env.IsFinished() {
			return nil
		}

		if err = m.ProcessRequest(env); err != nil {
			return err
		}
	}

	return nil
}

func (app *App) processResponse(env *context.Env) error {
	var err error
	for _, m := range app.middlewares {
		if env.IsFinished() {
			return nil
		}

		if err = m.ProcessResponse(env); err != nil {
			return err
		}
	}

	return nil
}
