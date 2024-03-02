package halo_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fengjx/go-halo/halo"
)

func TestContextWithValue(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "hello", "world")
	ctx = halo.WithValue(ctx, "hello1", "world1")
	assert.Equal(t, "world", ctx.Value("hello"))
	assert.Equal(t, "world1", ctx.Value("hello1"))
	assert.Equal(t, nil, ctx.Value("hello2"))
}

func TestContextSet(t *testing.T) {
	ctx := context.Background()
	ctx = halo.WithValue(ctx, "hello", "world")
	c := ctx.(*halo.Context)
	c.Set("hello1", "world1")
	assert.Equal(t, "world", ctx.Value("hello"))
	assert.Equal(t, "world1", ctx.Value("hello1"))
	assert.Equal(t, nil, ctx.Value("hello2"))
}

func TestContext(t *testing.T) {
	ctx := context.Background()
	ctx = halo.NewContext(ctx)
	doSomething(ctx)
	assert.Equal(t, "bar", ctx.Value("foo"))
}

func doSomething(ctx context.Context) {
	c := ctx.(*halo.Context)
	c.Set("foo", "bar")
}
