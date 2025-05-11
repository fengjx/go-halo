package errs

import (
	"errors"
	"fmt"
	"io"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/fengjx/go-halo/logger"
)

func TestWrap(t *testing.T) {
	err := errors.New("err1")
	err = Wrap(err, "err2")
	t.Log(fmt.Sprintf("v %v", err))
	t.Log(fmt.Sprintf("#v %#v", err))
	t.Log(fmt.Sprintf("+v %+v", err))

	logger.NewConsole().Info("log msg", zap.Error(err))
	logger.NewConsole(zap.AddCaller()).Info("log msg", zap.Error(err))
}

func TestCause(t *testing.T) {
	x := errors.New("error")
	tests := []struct {
		err  error
		want error
	}{{
		// nil error is nil
		err:  nil,
		want: nil,
	}, {
		// explicit nil error is nil
		err:  (error)(nil),
		want: nil,
	}, {
		// uncaused error is unaffected
		err:  io.EOF,
		want: io.EOF,
	}, {
		// caused error returns cause
		err:  Wrap(io.EOF, "ignored"),
		want: io.EOF,
	}, {
		err:  x, // return from errs.New
		want: x,
	}, {
		WithStack(nil),
		nil,
	}, {
		WithStack(io.EOF),
		io.EOF,
	}}

	for i, tt := range tests {
		got := Cause(tt.err)
		if !errors.Is(got, tt.want) {
			t.Errorf("test %d: got %#v, want %#v", i+1, got, tt.want)
		}
	}
}

func f1() error {
	log.Println("f1")
	return errors.New("f1 err")
}

func f2() error {
	log.Println("f2")
	err := f1()
	if err != nil {
		return WithStack(err)
	}
	return nil
}

func f3() error {
	log.Println("f3")
	err := f2()
	if err != nil {
		return WithStack(err)
	}
	return nil
}

func TestStack(t *testing.T) {
	err := f3()
	if err != nil {
		t.Logf("%+v", err)
		logger.NewConsole().Errorf("err log: %+v", err)
	}
}

func TestEquals(t *testing.T) {
	e1 := fmt.Errorf("error1: %w", io.EOF)
	e2 := fmt.Errorf("error2: %w", e1)
	e3 := Wrap(e2, "error3")
	t.Log("Unwrap 1", errors.Unwrap(e1))
	t.Log("Unwrap 2", errors.Unwrap(e2))
	t.Log("Unwrap 3", errors.Unwrap(e3))
	t.Log("Cause 1", Cause(e1))
	t.Log("Cause 2", Cause(e2))
	t.Log("Cause 3", Cause(e3))

	assert.True(t, errors.Is(e2, io.EOF))
	assert.True(t, errors.Is(e2, e1))
	assert.True(t, errors.Is(e3, e2))

	t.Logf("%v", e2)
}
