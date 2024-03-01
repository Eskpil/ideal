package ideal

import (
	"errors"
	"net/http"

	"github.com/graphql-go/graphql"
)

type key string

const RuntimeParametersKey key = "ideal.runtime_parameters"

type NextFunc func(ctx *Context) (interface{}, error)
type MiddlewareFunc func(ctx *Context, next NextFunc) (interface{}, error)

type ResolveFunc func(ctx *Context) (interface{}, error)

type handler struct {
	fieldName  string
	middleware []MiddlewareFunc
	resolve    ResolveFunc
}

type RuntimeParameters struct {
	Headers http.Header
	Cookies []*http.Cookie
}

type Runtime struct {
	handlers map[string]handler
}

func NewRuntime() *Runtime {
	return &Runtime{
		handlers: make(map[string]handler),
	}
}

func (r *Runtime) AddHandler(fieldName string, resolve ResolveFunc, middleware ...MiddlewareFunc) {
	r.handlers[fieldName] = handler{
		fieldName:  fieldName,
		middleware: middleware,
		resolve:    resolve,
	}
}

func (r *Runtime) Resolve(p graphql.ResolveParams) (interface{}, error) {
	params, ok := p.Context.Value(RuntimeParametersKey).(*RuntimeParameters)
	if !ok {
		return nil, errors.New("missing runtime parameters")
	}

	// we know it is safe because graphql will complain before us if the query/mutation/subscription does not exist.
	h := r.handlers[p.Info.FieldName]

	c := &Context{
		inner:     p.Context,
		p:         p,
		Arguments: p.Args,
		Headers:   params.Headers,
		Cookies:   params.Cookies,
	} // Create a new context here if necessary

	if h.middleware == nil {
		// If there is no middleware, execute the resolve function directly
		return h.resolve(c)
	}

	// Define the next function that will be passed to the middleware chain
	var index = 0
	var next NextFunc
	next = func(ctx *Context) (interface{}, error) {
		index++
		if index < len(h.middleware) {
			return h.middleware[index](c, next)
		}
		return h.resolve(c)
	}

	// Start the middleware chain with the first middleware
	return h.middleware[0](c, next)
}
