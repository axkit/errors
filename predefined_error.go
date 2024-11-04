package errors

// PredefinedError represents a custom error type that allows wrapping of other errors
// and includes additional attributes like context, severity, and status code.
type PredefinedError struct {
	// attrs contains the error's metadata (message, severity, status code, etc.).
	attrs

	// fields holds the error's custom key-value pairs.
	fields map[string]interface{}
}

// New creates a new predefined error with a specified message.
// This error type is useful for creating custom error types with predefined attributes.
//
// Example:
//
//	var ErrInvalidCustomerID := New("invalid customer ID").StatusCode(http.StatusBadRequest)
func New(msg string) *PredefinedError {
	return &PredefinedError{
		attrs: attrs{
			message: msg,
		},
	}
}

// Error returns the error message.
func (pe *PredefinedError) Error() string {
	return pe.message
}

func (pe *PredefinedError) toError() *Error {
	return &Error{
		attrs:  pe.attrs,
		fields: cloneMap(pe.fields),
	}
}

// Wrap wraps an error with a predefined error.
func (pe *PredefinedError) Wrap(err error) *Error {

	switch x := err.(type) {
	case *PredefinedError:
		return &Error{
			attrs:       pe.attrs,
			fields:      cloneMap(pe.fields),
			pureWrapper: true,
			err:         err,
			stack:       DefaultCallerFrames(1),
		}
	case *Error:
		res := Error{
			attrs:       pe.attrs,
			fields:      cloneMap(pe.fields),
			pureWrapper: true,
			err:         err,
		}
		if x.stack != nil {
			res.stack = x.stack
		} else {
			res.stack = DefaultCallerFrames(1)
		}
		if len(x.fields) > 0 {
			if res.fields == nil {
				res.fields = make(map[string]interface{}, len(x.fields))
			}
			for k, v := range x.fields {
				res.fields[k] = v
			}
		}
		return &res
	case error:
		return &Error{
			attrs:  pe.attrs,
			err:    err,
			fields: cloneMap(pe.fields),
			stack:  DefaultCallerFrames(1),
		}
	case nil:
		return &Error{
			attrs:  pe.attrs,
			fields: cloneMap(pe.fields),
			stack:  DefaultCallerFrames(1),
		}
	}
	panic("axkit/errors: invalid error type")
}

// Raise creates a new error instance with the predefined error's attributes.
func (pe *PredefinedError) Raise() *Error {
	return &Error{
		attrs:  pe.attrs,
		fields: cloneMap(pe.fields),
		stack:  CallerFramesFunc(1),
	}
}

// Set sets a custom key-value pair for the predefined error.
func (pe *PredefinedError) Set(key string, value any) *PredefinedError {
	if pe.fields == nil {
		pe.fields = make(map[string]any)
	}
	pe.fields[key] = value
	return pe
}

// Code sets an application-specific error code.
func (pe *PredefinedError) Code(code string) *PredefinedError {
	pe.code = code
	return pe
}

// Severity sets the severity level of the error.
func (pe *PredefinedError) Severity(severity SeverityLevel) *PredefinedError {
	pe.severity = severity
	return pe
}

// StatusCode sets the HTTP status code for the error.
func (pe *PredefinedError) StatusCode(statusCode int) *PredefinedError {
	pe.statusCode = statusCode
	return pe
}

// Protected marks the error as protected.
func (pe *PredefinedError) Protected(protected bool) *PredefinedError {
	pe.protected = protected
	return pe
}
