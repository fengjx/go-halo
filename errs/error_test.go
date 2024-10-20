package errs

import (
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"testing"

	"go.uber.org/zap"

	"github.com/fengjx/go-halo/logger"
)

func TestWrap(t *testing.T) {
	err := New("err1")
	err = Wrap(err, "err2")
	t.Log(fmt.Sprintf("v %v", err))
	t.Log(fmt.Sprintf("#v %#v", err))
	t.Log(fmt.Sprintf("+v %+v", err))

	logger.NewConsole().Info("log msg", zap.Error(err))
	logger.NewConsole(zap.AddCaller()).Info("log msg", zap.Error(err))
}

func TestCause(t *testing.T) {
	x := New("error")
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
		WithMessage(nil, "whoops"),
		nil,
	}, {
		WithMessage(io.EOF, "whoops"),
		io.EOF,
	}, {
		WithStack(nil),
		nil,
	}, {
		WithStack(io.EOF),
		io.EOF,
	}}

	for i, tt := range tests {
		got := Cause(tt.err)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("test %d: got %#v, want %#v", i+1, got, tt.want)
		}
	}
}

func TestWithMessage(t *testing.T) {
	tests := []struct {
		err     error
		message string
		want    string
	}{
		{io.EOF, "read error", "read error: EOF"},
		{WithMessage(io.EOF, "read error"), "client error", "client error: read error: EOF"},
	}

	for _, tt := range tests {
		got := WithMessage(tt.err, tt.message).Error()
		if got != tt.want {
			t.Errorf("WithMessage(%v, %q): got: %q, want %q", tt.err, tt.message, got, tt.want)
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
