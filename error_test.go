package errors

import (
	"errors"
	se "errors"
	"os"
	"testing"
)

func TestError(t *testing.T) {

	tcases := []struct {
		name       string
		msg        string
		err        error
		code       string
		statusCode int
		severity   SeverityLevel
		expected   string
	}{
		{
			name:       "plain error",
			msg:        "test error",
			err:        nil,
			code:       "X-0001",
			statusCode: 500,
			severity:   Critical,
			expected:   "test error",
		},
		{
			name:       "wrap predefined error",
			msg:        "test error",
			err:        errors.New("test error"),
			code:       "X-0001",
			statusCode: 500,
			severity:   Critical,
			expected:   "test error: test error",
		},
		{
			name:       "wrap os.ErrNotExist",
			msg:        "test error",
			err:        os.ErrNotExist,
			code:       "X-0001",
			statusCode: 500,
			severity:   Critical,
			expected:   "test error: file does not exist",
		},
		{
			name:       "wrap raised error",
			msg:        "test error",
			err:        Template("test error").New(),
			code:       "X-0001",
			statusCode: 500,
			severity:   Critical,
			expected:   "test error: test error",
		},
	}

	for _, tt := range tcases {
		t.Run(tt.name, func(t *testing.T) {
			pe1 := Template(tt.msg).
				Code(tt.code).
				StatusCode(tt.statusCode).
				Severity(tt.severity).Protected(true)
			err := pe1.Wrap(tt.err)

			pe2 := Template(tt.msg).Wrap(tt.err)
			pe2.Code(tt.code).
				StatusCode(tt.statusCode).
				Severity(tt.severity).Protected(true).
				Msg(tt.msg)

			if pe1.metadata != pe2.metadata {
				t.Errorf("expected %v, got %v", pe1.metadata, pe2.metadata)
			}

			if err.Error() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, err.Error())
			}

		})
	}
}

func TestError_WrappedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected int
	}{
		{
			name: "Single non-pure wrapper error",
			err: &Error{
				metadata:    metadata{message: "test error"},
				pureWrapper: false,
			},
			expected: 1,
		},
		{
			name: "Nested Error",
			err: &Error{
				metadata:    metadata{message: "test error"},
				pureWrapper: false,
				err: &Error{
					metadata:    metadata{message: "nested error"},
					pureWrapper: false,
				},
			},
			expected: 2,
		},
		{
			name: "Pure wrapper with nested error",
			err: &Error{
				pureWrapper: true,
				err: &Error{
					metadata:    metadata{message: "nested error"},
					pureWrapper: false,
				},
			},
			expected: 1,
		},
		{
			name: "Pure wrapper with nested pure wrapper",
			err: &Error{
				metadata: metadata{message: "nested error"},
				err:      os.ErrNotExist},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.WrappedErrors(); len(got) != tt.expected {
				t.Errorf("WrappedErrors() = %v, want %v", len(got), tt.expected)
			}
		})
	}
}

func TestError_Wrap(t *testing.T) {

	peTemplate := Template("predefined error").StatusCode(500).Protected(true)
	peTemplateWithFields := Template("predefined error with fields").StatusCode(500).Protected(true).Set("key", "value")
	e := Error{metadata: metadata{message: "test error"}}

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "Wrap predefined error",
			err:  peTemplate,
		},
		{
			name: "Wrap predefined error with fields",
			err:  peTemplateWithFields,
		},
		{
			name: "Wrap nil error",
			err:  nil,
		},
		{
			name: "Wrap standard error",
			err:  os.ErrNotExist,
		},
		{
			name: "Wrap raised predefined error with attributes",
			err:  Template("test error").Set("key1", "value1").New().StatusCode(500).Code("X-0001"),
		},
		{
			name: "Wrap raised error",
			err:  Template("test error").New(),
		},
		{
			name: "Wrap non-raised error",
			err:  e.Set("key1", "value1").StatusCode(500).Code("X-0001"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := peTemplate.New().Wrap(tt.err)
			if tt.err == nil {
				if !Is(err, err) {
					t.Errorf("expected wrapped error %v, got %v", tt.err, err.err)
				}
				return
			}

			if !Is(err, tt.err) {
				t.Errorf("expected wrapped error %v, got %v", tt.err, err.err)
			}
		})

		t.Run(tt.name+" with fields", func(t *testing.T) {
			err := peTemplateWithFields.New().Wrap(tt.err)
			if tt.err == nil {
				if !Is(err, err) {
					t.Errorf("expected wrapped error %v, got %v", tt.err, err.err)
				}
				return
			}

			if !Is(err, tt.err) {
				t.Errorf("expected wrapped error %v, got %v", tt.err, err.err)
			}
		})
	}
}
func TestError_PureWrapper(t *testing.T) {
	tests := []struct {
		name        string
		err         *Error
		expectedVal bool
	}{
		{
			name: "Pure wrapper is true",
			err: &Error{
				pureWrapper: true,
			},
			expectedVal: true,
		},
		{
			name: "Pure wrapper is false",
			err: &Error{
				pureWrapper: false,
			},
			expectedVal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.pureWrapper; got != tt.expectedVal {
				t.Errorf("pureWrapper = %v, want %v", got, tt.expectedVal)
			}
		})
	}
}

func TestError_Error(t *testing.T) {

	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "single non wrapping error",
			expected: "test error",
			err: &Error{
				metadata: metadata{message: "test error"},
			},
		},
		{
			name:     "wrapped error",
			expected: "outer error: inner error",
			err: &Error{
				metadata: metadata{message: "outer error"},
				err:      errors.New("inner error"),
			},
		},
		{
			name:     "wrapped standard error",
			expected: "outer error: file does not exist",
			err: &Error{
				metadata: metadata{message: "outer error"},
				err:      os.ErrNotExist,
			},
		},
		{
			name:     "wrapped standard error without message",
			expected: "outer error",
			err: &Error{
				metadata: metadata{message: "outer error"},
				err:      se.New(""),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestError_Alarm(t *testing.T) {
	mock := &MockAlarmer{}
	SetAlarmer(mock)

	testErr := Template("test error").New()
	testErr.Alarm()

	if !mock.called {
		t.Errorf("Expected Alarm to be called")
	}

	if mock.err != testErr {
		t.Errorf("Expected error to be %v, but got %v", testErr, mock.err)
	}
}
