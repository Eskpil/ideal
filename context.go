package ideal

import (
	"context"
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
)

type Context struct {
	inner context.Context
	p     graphql.ResolveParams

	Arguments map[string]interface{}

	Headers http.Header
	Cookies []*http.Cookie
}

func (c *Context) Source() any {
	return c.p.Source
}

func (c *Context) Cookie(name string) (*http.Cookie, error) {
	for _, cookie := range c.Cookies {
		if cookie.Name == name {
			return cookie, nil
		}
	}

	return nil, http.ErrNoCookie
}

func (c *Context) Header(key string) string {
	return c.Headers.Get(key)
}

func (c *Context) Set(key any, value any) {
	c.inner = context.WithValue(c.inner, key, value)
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.inner.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.inner.Done()
}

func (c *Context) Err() error {
	return c.inner.Err()
}

func (c *Context) Value(key any) any {
	return c.inner.Value(key)
}
