# polaris

a restful web framework like tornado written by go 

see [my chinese blog](http://blog.csdn.net/siddontang/article/details/21088451) for more

# Example

    type Handler1 struct {

    }

    func (h *Handler1) Prepare(env *Env) {
        fmt.Println("hello prepare")
    }

    func (h *Handler1) Get(env *Env) {
        env.WriteString("hello world")
    }

    type Handler2 struct {

    }

    //id is a captured submatch for regexp url below
    func (h *Handler2) Get(env *Env, id string) {
        env.WriteString("hello " + id)
    }

    r = NewRouter()

    r.Handle("/handler1", new(Handler1))
    r.Handle("/handler2/([0-9]+)", new(Handler2))

    http.Handle("/", r)
    http.ListenAndServe("127.0.0.1:11181", nil)