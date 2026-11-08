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

// New creates and returns a standard Go error using the built-in errors.New function.
//
// This is implemented to maintain compatibility with existing Go error handling practices.
// However, the design philosophy of this package does not encourage the use of errors.New(msg)
// as commonly practiced in Go. Instead, it promotes the use of structured and enriched error
// handling mechanisms provided by this package.
func New(msg string) error {
	return se.New(msg)
}

// ErrUnsupported is the standard errors.ErrUnsupported, so errors.Is
// works the same with both packages.
var ErrUnsupported = se.ErrUnsupported

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error. Otherwise, Unwrap returns nil.
// It is the standard errors.Unwrap.
func Unwrap(err error) error {
	return se.Unwrap(err)
}

// Join returns an error that wraps the given errors. Nil errors are discarded.
// Join returns nil if every value in errs is nil.
// It is the standard errors.Join.
func Join(errs ...error) error {
	return se.Join(errs...)
}

// AsType finds the first error in err's chain that matches the type E,
// and if one is found, returns that error value and true.
// It works the same way as As and mirrors errors.AsType added in Go 1.26.
func AsType[E error](err error) (E, bool) {
	var target E
	ok := As(err, &target)
	return target, ok
}

// Wrap wraps an existing error with a new message, effectively creating
// a new error that includes the previous error.
//
// If err is an *Error or *ErrorTemplate, the new error inherits its severity,
// status code, code, protected flag, fields and stack trace. The wrapped error
// stays in the chain, so Error() returns "message: <err.Error()>".
// The wrapped error itself is not modified.
func Wrap(err error, message string) *Error {
	res := Error{err: err}

	switch x := err.(type) {
	case *Error:
		if x != nil {
			res.metadata = x.metadata
			res.fields = cloneMap(x.fields)
			res.stack = x.stack
		}
	case *ErrorTemplate:
		if x != nil {
			res.metadata = x.metadata
			res.fields = cloneMap(x.fields)
		}
	}

	res.message = message
	if len(res.stack) == 0 {
		res.stack = CallerFramesFunc(callerFramesSkip)
	}

	return &res
}

// Is reports whether any error in err's chain matches target.
//
// An *Error matches an *ErrorTemplate if it was created from the template
// by New or Wrap, even if Msg, StatusCode and others changed its attributes,
// or if all attributes (message, severity, status code, code, protected) are equal.
// Two *Error values match if they are the same, were created from the same
// template or have equal attributes. Other errors are compared as by the standard errors.Is.
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
		if e == t || e.metadata.equal(t.metadata) || sameTemplate(e, t) {
			return true
		}
	case *ErrorTemplate:
		if e.tpl == t || t.metadata.equal(e.metadata) {
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

// As finds the first error in err's chain that matches target, and if one is found,
// sets target to that error value and returns true.
//
// If target is a pointer to a nil *Error, any *Error matches. If it points to
// a non-nil *Error, the error must match it by the rules of Is.
// Other targets are handled as by the standard errors.As.
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
		// A nil target matches any *Error, as the standard errors.As does.
		if *t == nil {
			*t = e
			return true
		}

		if e == *t || e.metadata.equal((*t).metadata) || sameTemplate(e, *t) {
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

// sameTemplate reports whether both errors were created from the same template.
func sameTemplate(a, b *Error) bool {
	return a.tpl != nil && a.tpl == b.tpl
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
