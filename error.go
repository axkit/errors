package errors

type Error struct {
	// ctx holds the context for the error, which can be used to store additional key-value pairs.
	fields map[string]interface{}
	stack  []StackFrame
	attrs

	pureWrapper bool
	err         error
}

// Error returns the error message.
func (e *Error) Error() string {

	res := e.message

	if e.err != nil {
		child := e.err.Error()
		if child != "" {
			if res != "" {
				res += ": " + child
			} else {
				res = child
			}
		}
	}

	return res
}

// WrappedErrors returns all the errors that have been wrapped by this custom error, including itself.
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
			if !x.attrs.empty() {
				res = append(res, *x)
			}
			e = x.err
		case *PredefinedError:
			res = append(res, *x.toError())
			e = nil
		default:
			res = append(res, Error{err: e})
			e = nil
		}
	}
	return res
}

// Wrap wraps an error with a predefined error.
func (e *Error) Wrap(err error) *Error {

	if err == nil {
		return e
	}

	switch x := err.(type) {
	case *PredefinedError:
		return &Error{
			attrs:       e.attrs,
			fields:      cloneMap(e.fields),
			pureWrapper: true,
			err:         err,
			stack:       DefaultCallerFrames(1),
		}
	case *Error:
		res := Error{
			attrs:       e.attrs,
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
				res.fields = make(map[string]interface{}, len(x.fields))
			}
			for k, v := range x.fields {
				res.fields[k] = v
			}
		}
		return &res
	}

	return &Error{
		attrs:       e.attrs,
		err:         err,
		fields:      cloneMap(e.fields),
		pureWrapper: true,
		stack:       DefaultCallerFrames(1),
	}
}

// Set sets a custom key-value pair for the predefined error.
func (e *Error) Set(key string, value any) *Error {
	if e.fields == nil {
		e.fields = make(map[string]any)
	}
	e.fields[key] = value
	return e
}

// Code sets an application-specific error code.
func (e *Error) Code(code string) *Error {
	e.code = code
	return e
}

// Severity sets the severity level of the error.
func (e *Error) Severity(severity SeverityLevel) *Error {
	e.severity = severity
	return e
}

// StatusCode sets the HTTP status code for the error.
func (e *Error) StatusCode(statusCode int) *Error {
	e.statusCode = statusCode
	return e
}

// Protected marks the error as protected. Protected errors could be used to avoid certain types of modifications or exposure.
func (e *Error) Protected(protected bool) *Error {
	e.protected = protected
	return e
}

// Msg sets the error message. It overrides the existing message.
func (e *Error) Msg(s string) *Error {
	e.message = s
	e.pureWrapper = false
	return e
}

// Alarm sends an alert for the error if an alarmer is set.
func (e *Error) Alarm() {
	if alarmer != nil {
		alarmer.Alarm(e)
	}
}
