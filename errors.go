package errors

import (
	se "errors"
)

// attrs defines additional metadata that can be attached to an error.
type attrs struct {
	// message holds the final error's message.
	message string

	// severity holds the severity level of the error.
	severity SeverityLevel

	// statusCode holds the HTTP status code recommended for an HTTP response if specified.
	statusCode int

	// code holds the application-specific error code.
	code string

	// protected marks the error as protected to avoid certain types of modifications or exposure.
	protected bool
}

func (a *attrs) empty() bool {
	return a.message == "" && a.severity == 0 && a.statusCode == 0 && a.code == "" && !a.protected
}

// Wrap wraps an existing error with a new message, effectively creating
// a new error that includes the previous error.
func Wrap(err error, message string) *Error {
	res := Error{
		err: err,
	}

	switch x := err.(type) {
	case *Error:
		res = *x
		x.pureWrapper = false
	case *PredefinedError:
		res.attrs = x.attrs
		res.fields = cloneMap(x.fields)
	case error:
		break
	default:
		break
	}

	res.message = message
	if len(res.stack) == 0 {
		res.stack = CallerFramesFunc(1)
	}

	return &res
}

func cloneMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}

	clone := make(map[string]interface{}, len(m))
	for k, v := range m {
		clone[k] = v
	}
	return clone
}

// Is checks if the error is of the same type as the target error.
func Is(err error, target error) bool {
	if err == nil || target == nil {
		return err == target
	}

	switch x := err.(type) {
	case *Error:
		return is(x, target)
	case *PredefinedError:
		return is(x.toError(), target)
	}

	return se.Is(err, target)
}

// is checks if two custom errors are equal based on their attributes or if their wrapped errors are equal.
func is(e *Error, target error) bool {
	switch t := target.(type) {
	case *Error:
		if t.pureWrapper {
			return is(e, t.err)
		}
		if e.attrs == t.attrs || e == t {
			return true
		}
	case *PredefinedError:
		if t.attrs == e.attrs {
			return true
		}
	default:
		return se.Is(e.err, target)
	}

	if e.err == nil {
		return false
	}

	switch x := e.err.(type) {
	case *Error:
		return is(x, target)
	case *PredefinedError:
		return is(x.toError(), target)
	}

	return se.Is(e.err, target)
}

// As checks if the error can be cast to a target type.
func As(err error, target any) bool {
	if err == nil {
		return false
	}

	if target == nil {
		panic("axkit/errors: target cannot be nil")
	}

	if e, ok := err.(*Error); ok {
		return as(e, target)
	}

	if _, ok := err.(*PredefinedError); ok {
		panic("axkit/errors: error cannot be a pointer to a PredefinedError")
	}

	return se.As(err, target)
}

// as assists the As function in casting errors to the target type, accounting for wrapped errors.
func as(e *Error, target any) bool {

	switch t := target.(type) {
	case **Error:
		if e.attrs == (*t).attrs || e == *t {
			*t = e
			return true
		}

		if e.err == nil {
			return false
		}

		switch x := e.err.(type) {
		case *Error:
			return as(x, target)
		case *PredefinedError:
			if e.attrs == x.attrs {
				*t = (*x).toError()
				return true
			}
		}
	case **PredefinedError, *PredefinedError:
		panic("axkit/errors: target cannot be a pointer to a PredefinedError")
	}

	return se.As(e, target)
}
