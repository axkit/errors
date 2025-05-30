package errors

import (
	"fmt"
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
