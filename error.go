package errors

// Error represents a structured error with metadata, custom fields, stack trace, and optional wrapping.
type Error struct {
	metadata
	fields map[string]any
	stack  []StackFrame

	pureWrapper bool
	err         error
}

// Error returns the error message, including any wrapped error messages.
func (e *Error) Error() string {

	res := e.message

	if e.err != nil {
		prev := e.err.Error()
		if prev != "" {
			if res != "" {
				res += ": " + prev
			} else {
				res = prev
			}
		}
	}

	return res
}

// WrappedErrors returns a slice of all wrapped errors, including the current one if it's not a pure wrapper.
func (err *Error) WrappedErrors() []Error {
	var res []Error
	if !err.pureWrapper {
		res = []Error{*err}
	}

	e := err.err
	for {
		if e == nil {
			break
		}

		switch x := e.(type) {
		case *Error:
			if !x.metadata.empty() {
				res = append(res, *x)
			}
			e = x.err
		case *ErrorTemplate:
			res = append(res, *x.toError())
			e = nil
		default:
			res = append(res, Error{err: e})
			e = nil
		}
	}
	return res
}

// Wrap returns a new Error that wraps the given error while retaining the current error's metadata and fields.
// It also preserves the stack trace of the wrapped error if available.
// The new error is marked as a pure wrapper if the original error is of type ErrorTemplate or Error.
// If the original error is nil, it returns the current error.
// This method is useful for chaining errors and maintaining context.
// It supports wrapping both ErrorTemplate and Error types, preserving their fields and stack trace.
func (e *Error) Wrap(err error) *Error {

	if err == nil {
		return e
	}

	switch x := err.(type) {
	case *ErrorTemplate:
		return &Error{
			metadata:    e.metadata,
			fields:      cloneMap(e.fields),
			pureWrapper: true,
			err:         err,
			stack:       DefaultCallerFrames(1),
		}
	case *Error:
		res := Error{
			metadata:    e.metadata,
			fields:      cloneMap(e.fields),
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
				res.fields = make(map[string]any, len(x.fields))
			}
			for k, v := range x.fields {
				res.fields[k] = v
			}
		}
		return &res
	}

	return &Error{
		metadata:    e.metadata,
		err:         err,
		fields:      cloneMap(e.fields),
		pureWrapper: true,
		stack:       DefaultCallerFrames(1),
	}
}

// Set adds or updates a custom key-value pair in the error's fields.
func (e *Error) Set(key string, value any) *Error {
	if e.fields == nil {
		e.fields = make(map[string]any)
	}
	e.fields[key] = value
	return e
}

// Code sets a custom application-specific code for the error.
func (e *Error) Code(code string) *Error {
	e.code = code
	return e
}

// Severity sets the severity level for the error.
func (e *Error) Severity(severity SeverityLevel) *Error {
	e.severity = severity
	return e
}

// StatusCode sets the associated HTTP status code for the error.
func (e *Error) StatusCode(statusCode int) *Error {
	e.statusCode = statusCode
	return e
}

// Protected marks the error as protected to prevent certain modifications or exposure.
func (e *Error) Protected(protected bool) *Error {
	e.protected = protected
	return e
}

// Msg sets the error message and marks the error as not being a pure wrapper.
func (e *Error) Msg(s string) *Error {
	e.message = s
	e.pureWrapper = false
	return e
}

// Alarm triggers an alert for the error if an alarmer is configured.
func (e *Error) Alarm() {
	if alarmer != nil {
		alarmer.Alarm(e)
	}
}
