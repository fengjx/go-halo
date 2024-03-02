package halo

import (
	"context"
	"time"
)

type Context struct {
	ctx context.Context
}

func NewContext(ctx context.Context) *Context {
	return &Context{ctx: ctx}
}

func WithValue(parent context.Context, key, val any) *Context {
	ctx := context.WithValue(parent, key, val)
	return &Context{ctx: ctx}
}

func (c *Context) Set(key, val any) {
	c.ctx = context.WithValue(c.ctx, key, val)
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *Context) Err() error {
	return c.ctx.Err()
}

func (c *Context) Value(key any) any {
	return c.ctx.Value(key)
}
