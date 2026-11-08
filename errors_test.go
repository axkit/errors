package errors

import (
	se "errors"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	msg := "test error"
	err := New(msg)
	if err == nil {
		t.Errorf("expected non-nil error, got nil")
	}

	if err.Error() != msg {
		t.Errorf("expected error message %q, got %q", msg, err.Error())
	}
}

func TestWrap(t *testing.T) {
	var err = Template("test error")
	var wrappedBaseErr = fmt.Errorf("wrapped error: %w", os.ErrNotExist)

	test := []struct {
		name     string
		err      error
		message  string
		expected string
	}{
		{
			name:     "Wrap error",
			err:      err,
			message:  "wrapped error",
			expected: "wrapped error: test error",
		},
		{
			name:     "Wrap nil error",
			err:      nil,
			message:  "wrapped error",
			expected: "wrapped error",
		},
		{
			name:     "Wrap standard error",
			err:      wrappedBaseErr,
			message:  "wrapped error",
			expected: "wrapped error: wrapped error: file does not exist",
		},

		{
			name:     "Wrap nil standard error",
			err:      nil,
			message:  "wrapped error",
			expected: "wrapped error",
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := Wrap(tt.err, tt.message)
			if wrapped.Error() != tt.expected {
				t.Errorf("Wrap() = %v, want %v", wrapped.Error(), tt.expected)
			}
		})
	}
}

func TestIs(t *testing.T) {

	var err = Template("test error")
	var wrappedBaseErr = fmt.Errorf("wrapped error: %w", os.ErrNotExist)
	tests := []struct {
		name     string
		err      error
		target   error
		expected bool
	}{
		{
			name:     "Same error",
			err:      err,
			target:   err,
			expected: true,
		},
		{
			name:     "Wrapped error",
			err:      Wrap(err, "wrapped error"),
			target:   err,
			expected: true,
		},
		{
			name:     "Different error",
			err:      Template("different error"),
			target:   err,
			expected: false,
		},
		{
			name:     "Nil error",
			err:      nil,
			target:   err,
			expected: false,
		},
		{
			name:     "Nil target",
			err:      err,
			target:   nil,
			expected: false,
		},
		{
			name:     "standard error",
			err:      wrappedBaseErr,
			target:   os.ErrNotExist,
			expected: true,
		},
		{
			name: "pure wrapper error",
			err: &Error{
				pureWrapper: true,
				err:         os.ErrNotExist,
			},
			target: &Error{
				pureWrapper: true,
				err:         os.ErrNotExist,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if Is(tt.err, tt.target) != tt.expected {
				t.Errorf("Is() = %v, want %v", Is(tt.err, tt.target), tt.expected)
			}
		})
	}

}

func TestAs(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		if As(nil, os.ErrNotExist) {
			t.Errorf("expected false, got %v", As(nil, nil))
		}
	})
	t.Run("nil target", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic, did not catch it")
			}
		}()
		As(Template("test error"), nil)
	})

	t.Run("err and target are predefined errors", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic, did not catch it")
			}
		}()
		As(Template("test error"), Template("test error"))
	})

	t.Run("standard error", func(t *testing.T) {
		err := os.ErrNotExist
		if !As(os.ErrNotExist, &err) {
			t.Errorf("expected true, got false")
		}
	})

	t.Run("with Error", func(t *testing.T) {
		err := Template("test error").New().Set("key", "value")
		target := Template("test error").New()
		if !As(err, &target) {
			t.Errorf("expected true, got false")
		}
		if target.fields["key"] != "value" {
			t.Errorf("expected field to be %q, got %q", "value", target.fields["key"])
		}
	})
	t.Run("with wrapped Error", func(t *testing.T) {
		err := Template("inner error").New().Wrap(os.ErrNotExist)
		werr := Template("outer error").New().Wrap(err)

		if !As(werr, &err) {
			t.Errorf("expected true, got false")
		}
	})
	t.Run("with wrapped predefined error", func(t *testing.T) {
		pe := Template("predefined error")
		err := Template("outer error").New().Wrap(pe)

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic, did not catch it")
			}
		}()
		As(err, &pe)
	})
}

func TestAsNilTarget(t *testing.T) {
	src := Template("test error").Code("C1").New()

	t.Run("direct", func(t *testing.T) {
		var target *Error
		if !As(src, &target) {
			t.Fatal("expected true, got false")
		}
		if target != src {
			t.Errorf("expected target to be the source error")
		}
	})

	t.Run("through fmt.Errorf", func(t *testing.T) {
		var target *Error
		if !As(fmt.Errorf("ctx: %w", src), &target) {
			t.Fatal("expected true, got false")
		}
		if target != src {
			t.Errorf("expected target to be the source error")
		}
	})
}

func TestStdErrorsCompatibility(t *testing.T) {
	tpl := Template("not found").Code("NF")
	other := Template("other").Code("OT")
	pathErr := &os.PathError{Op: "open", Path: "/x", Err: os.ErrNotExist}

	t.Run("Unwrap", func(t *testing.T) {
		e := tpl.Wrap(io.EOF)
		if e.Unwrap() != io.EOF {
			t.Errorf("expected io.EOF, got %v", e.Unwrap())
		}
		if (&Error{}).Unwrap() != nil {
			t.Error("expected nil for error without wrapped error")
		}
	})

	t.Run("std Is finds template through fmt.Errorf", func(t *testing.T) {
		err := fmt.Errorf("ctx: %w", tpl.New())
		if !se.Is(err, tpl) {
			t.Error("expected true, got false")
		}
		if se.Is(err, other) {
			t.Error("expected false for other template")
		}
		if !Is(err, tpl) {
			t.Error("package Is: expected true, got false")
		}
	})

	t.Run("std Is finds wrapped std error", func(t *testing.T) {
		err := fmt.Errorf("ctx: %w", tpl.Wrap(io.EOF))
		if !se.Is(err, io.EOF) {
			t.Error("expected true, got false")
		}
	})

	t.Run("std Is finds inner template", func(t *testing.T) {
		err := other.Wrap(tpl.New())
		if !se.Is(err, tpl) {
			t.Error("expected true, got false")
		}
	})

	t.Run("As finds std error inside Error", func(t *testing.T) {
		err := tpl.Wrap(pathErr)
		var target *os.PathError
		if !As(err, &target) || target != pathErr {
			t.Errorf("package As: expected %v, got %v", pathErr, target)
		}
		target = nil
		if !se.As(err, &target) || target != pathErr {
			t.Errorf("std As: expected %v, got %v", pathErr, target)
		}
	})
}

func TestWrapChain(t *testing.T) {
	errNotFound := Template("record not found").Code("DB-404").StatusCode(404)
	errUserLoad := Template("failed to load user").Code("USR-500").StatusCode(500)

	t.Run("template, template, package Wrap", func(t *testing.T) {
		e1 := errNotFound.Wrap(io.EOF)
		e2 := errUserLoad.Wrap(e1)
		e3 := Wrap(e2, "handle request")

		want := "handle request: failed to load user: record not found: EOF"
		if e3.Error() != want {
			t.Errorf("Error() = %q, want %q", e3.Error(), want)
		}
		if e3.code != "USR-500" || e3.statusCode != 500 {
			t.Errorf("expected metadata of the wrapped error, got code=%q status=%d", e3.code, e3.statusCode)
		}
		if !Is(e3, errUserLoad) || !Is(e3, errNotFound) || !Is(e3, io.EOF) {
			t.Error("expected all errors of the chain to match")
		}
	})

	t.Run("package Wrap three times", func(t *testing.T) {
		w := Wrap(Wrap(Wrap(io.EOF, "read header"), "parse config"), "start app")
		want := "start app: parse config: read header: EOF"
		if w.Error() != want {
			t.Errorf("Error() = %q, want %q", w.Error(), want)
		}
	})

	t.Run("wrapped error is not modified", func(t *testing.T) {
		inner := errNotFound.New().Set("id", 1)
		before := *inner
		outer := Wrap(inner, "outer").Set("id", 2).Set("extra", true)

		if inner.message != before.message || inner.pureWrapper != before.pureWrapper || inner.err != before.err {
			t.Errorf("inner error was modified: %+v", inner)
		}
		if len(inner.fields) != 1 || inner.fields["id"] != 1 {
			t.Errorf("inner fields were modified: %v", inner.fields)
		}
		if outer.fields["id"] != 2 {
			t.Errorf("expected outer field id=2, got %v", outer.fields["id"])
		}
	})

	t.Run("keeps original stack", func(t *testing.T) {
		inner := errNotFound.New()
		if outer := Wrap(inner, "outer"); &outer.stack[0] != &inner.stack[0] {
			t.Error("expected the stack of the wrapped error")
		}
	})

	t.Run("inherits protected flag", func(t *testing.T) {
		if !Wrap(Template("sql").Protected(true).New(), "load").protected {
			t.Error("expected protected flag to be inherited")
		}
	})

	t.Run("nil *Error", func(t *testing.T) {
		var e *Error
		if w := Wrap(e, "msg"); w.message != "msg" {
			t.Errorf("unexpected result: %+v", w)
		}
	})
}

func TestIsByTemplate(t *testing.T) {
	tpl := Template("invalid input").Code("CRM-0901").StatusCode(400)
	twin := Template("invalid input").Code("CRM-0901").StatusCode(400)
	other := Template("other").Code("OTH")

	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{"New().Msg", tpl.New().Msg("empty email"), tpl, true},
		{"New().StatusCode", tpl.New().StatusCode(422), tpl, true},
		{"New().Code().Severity()", tpl.New().Code("X").Severity(Critical), tpl, true},
		{"Wrap().Msg", tpl.Wrap(io.EOF).Msg("x"), tpl, true},
		{"Wrap of changed error", other.Wrap(tpl.New().Msg("x")), tpl, true},
		{"package Wrap of changed error", Wrap(tpl.New().Msg("x"), "ctx"), tpl, true},
		{"fmt wrapper of changed error", fmt.Errorf("ctx: %w", tpl.New().Msg("x")), tpl, true},
		{"instance Wrap keeps template", tpl.New().Msg("x").Wrap(io.EOF), tpl, true},
		{"two changed instances", tpl.New().Msg("a"), tpl.New().Msg("b"), true},
		{"twin template by attributes", tpl.New(), twin, true},
		{"twin template after Msg", tpl.New().Msg("x"), twin, false},
		{"other template", tpl.New().Msg("x"), other, false},
		{"instances of different templates", tpl.New().Msg("x"), other.New(), false},
		{"error without template", &Error{metadata: metadata{message: "x"}}, tpl, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Is(tt.err, tt.target); got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
			if got := se.Is(tt.err, tt.target); got != tt.want {
				t.Errorf("std errors.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAsByTemplate(t *testing.T) {
	tpl := Template("invalid input").Code("CRM-0901")
	other := Template("other")
	src := tpl.New().Msg("empty email")

	t.Run("same template", func(t *testing.T) {
		target := tpl.New()
		if !As(fmt.Errorf("ctx: %w", src), &target) || target != src {
			t.Errorf("expected source error, got %+v", target)
		}
	})

	t.Run("other template", func(t *testing.T) {
		target := other.New()
		if As(src, &target) {
			t.Error("expected false")
		}
	})
}
