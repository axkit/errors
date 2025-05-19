// Package errors provides a structured and extensible way to create, wrap, and manage errors
// in Go applications. It includes support for adding contextual information, managing error
// hierarchies, and setting attributes such as severity, HTTP status codes, and custom error codes.
//
// The package is designed to enhance error handling by allowing developers to attach additional
// metadata to errors, wrap underlying errors with more context, and facilitate debugging and
// logging. It also supports integration with alerting systems through the Alarm method.
//
// Key features include:
// - Wrapping errors with additional context.
// - Setting custom attributes like severity, status codes, and business codes.
// - Managing error stacks and hierarchies.
// - Sending alerts for critical errors.
// - Support for custom key-value pairs to enrich error information.
// - Integration with predefined error types for common scenarios.
// - Serialization errors for easy logging.
package errors

import (
	se "errors"
	"reflect"
)

// metadata holds the metadata for an error, including its message, severity level, etc.
type metadata struct {
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

func (a *metadata) empty() bool {
	return a.message == "" &&
		a.severity == 0 &&
		a.statusCode == 0 && a.code == "" &&
		!a.protected
}

func (a *metadata) equal(b metadata) bool {
	return a.message == b.message &&
		a.severity == b.severity &&
		a.statusCode == b.statusCode && a.code == b.code &&
		a.protected == b.protected
}

// Wrap wraps an existing error with a new message, effectively creating
// a new error that includes the previous error.
func Wrap(err error, message string) *Error {
	var res Error

	if err != nil {
		res.err = err
		switch x := err.(type) {
		case *Error:
			res = *x
			x.pureWrapper = false
		case *ErrorTemplate:
			res.metadata = x.metadata
			res.fields = cloneMap(x.fields)
		case error:
			break
		default:
			break
		}
	}

	res.message = message
	if len(res.stack) == 0 {
		res.stack = CallerFramesFunc(1)
	}

	return &res
}

// Is checks if the error is of the same type as the target error.
func Is(err error, target error) bool {
	if err == target {
		return true
	}

	if err == nil || target == nil {
		return err == target
	}

	switch x := err.(type) {
	case *Error:
		return is(x, target)
	case *ErrorTemplate:
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
		if e.metadata.equal(t.metadata) || e == t {
			return true
		}
	case *ErrorTemplate:
		if t.metadata.equal(e.metadata) {
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
	case *ErrorTemplate:
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

	if reflect.TypeOf(target).Kind() != reflect.Ptr {
		panic("axkit/errors: target must be a non-nil pointer")
	}

	if e, ok := err.(*Error); ok {
		return as(e, target)
	}

	if _, ok := err.(*ErrorTemplate); ok {
		panic("axkit/errors: error cannot be a pointer to a ErrorTemplate")
	}

	return se.As(err, target)
}

// as assists the As function in casting errors to the target type, accounting for wrapped errors.
func as(e *Error, target any) bool {

	switch t := target.(type) {
	case **Error:
		if e.metadata.equal((*t).metadata) || e == *t {
			*t = e
			return true
		}

		if e.err == nil {
			return false
		}

		switch x := e.err.(type) {
		case *Error:
			return as(x, target)
		case *ErrorTemplate:
			if e.metadata.equal(x.metadata) {
				*t = (*x).toError()
				return true
			}
		}
	case **ErrorTemplate, *ErrorTemplate:
		panic("axkit/errors: target cannot be a pointer to a PredefinedError")
	}

	return se.As(e, target)
}

func cloneMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}

	clone := make(map[string]any, len(m))
	for k, v := range m {
		clone[k] = v
	}
	return clone
}
