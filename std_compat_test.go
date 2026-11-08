package errors

import (
	se "errors"
	"fmt"
	"io"
	"os"
	"testing"
)

// The package must provide the whole API of the standard errors package
// with the same signatures, so it can replace the import.
var (
	_ func(string) error         = New
	_ func(error, error) bool    = Is
	_ func(error, any) bool      = As
	_ func(error) error          = Unwrap
	_ func(...error) error       = Join
	_ error                      = ErrUnsupported
	_ func(error) (*Error, bool) = AsType[*Error]
)

func TestErrUnsupported(t *testing.T) {
	if ErrUnsupported != se.ErrUnsupported {
		t.Error("expected the standard ErrUnsupported value")
	}
	if !Is(fmt.Errorf("op: %w", se.ErrUnsupported), ErrUnsupported) {
		t.Error("expected Is to match")
	}
}

func TestUnwrap(t *testing.T) {
	inner := Template("inner").New()
	tests := []struct {
		name string
		err  error
		want error
	}{
		{"nil", nil, nil},
		{"std error", io.EOF, nil},
		{"fmt wrapper", fmt.Errorf("ctx: %w", io.EOF), io.EOF},
		{"Error", Template("tpl").Wrap(inner), inner},
		{"package Wrap", Wrap(io.EOF, "msg"), io.EOF},
		{"template", Template("tpl"), nil},
		{"join is not unwrapped", Join(io.EOF), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Unwrap(tt.err); got != tt.want {
				t.Errorf("Unwrap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	tplA := Template("a").Code("A")
	tplB := Template("b").Code("B")
	tplC := Template("c").Code("C")

	t.Run("nil", func(t *testing.T) {
		if Join() != nil || Join(nil, nil) != nil {
			t.Error("expected nil")
		}
	})

	err := Join(tplA.New(), tplB.Wrap(io.EOF))

	t.Run("message", func(t *testing.T) {
		if want := "a\nb: EOF"; err.Error() != want {
			t.Errorf("Error() = %q, want %q", err.Error(), want)
		}
	})

	t.Run("Is", func(t *testing.T) {
		for _, target := range []error{tplA, tplB, io.EOF} {
			if !Is(err, target) || !se.Is(err, target) {
				t.Errorf("expected %v to match", target)
			}
		}
		if Is(err, tplC) || se.Is(err, tplC) {
			t.Error("expected other template not to match")
		}
	})

	t.Run("As", func(t *testing.T) {
		var e *Error
		if !As(err, &e) || e.code != "A" {
			t.Errorf("expected first *Error, got %+v", e)
		}
		pathErr := &os.PathError{Op: "open", Path: "/x", Err: os.ErrNotExist}
		var pe *os.PathError
		if !As(Join(io.EOF, tplA.Wrap(pathErr)), &pe) || pe != pathErr {
			t.Errorf("expected %v, got %v", pathErr, pe)
		}
	})
}

func TestAsType(t *testing.T) {
	pathErr := &os.PathError{Op: "open", Path: "/x", Err: os.ErrNotExist}

	t.Run("std error type", func(t *testing.T) {
		pe, ok := AsType[*os.PathError](Template("tpl").Wrap(pathErr))
		if !ok || pe != pathErr {
			t.Errorf("got %v, %v", pe, ok)
		}
	})

	t.Run("Error", func(t *testing.T) {
		src := Template("tpl").New()
		e, ok := AsType[*Error](fmt.Errorf("ctx: %w", src))
		if !ok || e != src {
			t.Errorf("got %v, %v", e, ok)
		}
	})

	t.Run("not found", func(t *testing.T) {
		pe, ok := AsType[*os.PathError](io.EOF)
		if ok || pe != nil {
			t.Errorf("got %v, %v", pe, ok)
		}
	})

	t.Run("nil error", func(t *testing.T) {
		if _, ok := AsType[*Error](nil); ok {
			t.Error("expected false")
		}
	})
}

func TestToJSONProtectedInJoin(t *testing.T) {
	errSQL := Template("pq: secret").Protected(true)
	errPublic := Template("public")

	tests := []struct {
		name string
		err  error
		want string
	}{
		{"protected in second branch skipped",
			errPublic.Wrap(Join(io.EOF, errSQL.New())),
			`{"msg":"public","severity":"unknown"}`},
		{"protected template in join skipped",
			errPublic.Wrap(Join(errSQL, io.EOF)),
			`{"msg":"public","severity":"unknown"}`},
		{"protected deep in fmt and join skipped",
			errPublic.Wrap(fmt.Errorf("ctx: %w", Join(io.EOF, Wrap(errSQL.New(), "load")))),
			`{"msg":"public","severity":"unknown"}`},
		{"public join written",
			errPublic.Wrap(Join(io.EOF, Template("x").New())),
			`{"msg":"public","severity":"unknown","wrapped":[{"msg":"EOF\nx","severity":"unknown"}]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jsonString(tt.err, AddWrappedErrors); got != tt.want {
				t.Errorf("got  %s\nwant %s", got, tt.want)
			}
		})
	}
}
