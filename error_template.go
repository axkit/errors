package errors

// ErrorTemplate defines a reusable error blueprint that includes metadata
// and custom key-value fields. It is designed for creating structured errors
// with consistent attributes such as severity and HTTP status code.
type ErrorTemplate struct {
	// metadata contains the error's metadata (message, severity, status code, etc.).
	metadata

	// fields holds the error's custom key-value pairs.
	fields map[string]any
}

// Template returns a new ErrorTemplate initialized with the given message.
// It can be extended with additional attributes and reused to create
// multiple error instances.
func Template(msg string) *ErrorTemplate {
	return &ErrorTemplate{
		metadata: metadata{
			message: msg,
		},
	}
}

// Error returns the error message from the template.
func (et *ErrorTemplate) Error() string {
	return et.message
}

// toError converts the ErrorTemplate to an Error instance.
func (et *ErrorTemplate) toError() *Error {
	return &Error{
		metadata: et.metadata,
		fields:   cloneMap(et.fields),
	}
}

// Wrap wraps an existing error with the ErrorTemplate's metadata and fields.
// It supports wrapping both ErrorTemplate and Error types,
// preserving their fields and stack trace.
func (et *ErrorTemplate) Wrap(err error) *Error {

	switch x := err.(type) {
	case *ErrorTemplate:
		return &Error{
			metadata:    et.metadata,
			fields:      cloneMap(et.fields),
			pureWrapper: true,
			err:         err,
			stack:       DefaultCallerFrames(1),
		}
	case *Error:
		res := Error{
			metadata:    et.metadata,
			fields:      cloneMap(et.fields),
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
		metadata:    et.metadata,
		pureWrapper: true,
		err:         err,
		fields:      cloneMap(et.fields),
		stack:       DefaultCallerFrames(1),
	}
}

// New creates a new Error instance using the template's metadata and fields.
// A new stack trace is captured at the point of the call.
func (et *ErrorTemplate) New() *Error {
	return &Error{
		metadata: et.metadata,
		fields:   cloneMap(et.fields),
		stack:    CallerFramesFunc(1),
	}
}

// Set adds a custom key-value pair to the template's fields.
func (et *ErrorTemplate) Set(key string, value any) *ErrorTemplate {
	if et.fields == nil {
		et.fields = make(map[string]any)
	}
	et.fields[key] = value
	return et
}

// Code sets an application-specific error code on the template.
func (et *ErrorTemplate) Code(code string) *ErrorTemplate {
	et.code = code
	return et
}

// Severity sets the severity level for the error template.
func (et *ErrorTemplate) Severity(severity SeverityLevel) *ErrorTemplate {
	et.severity = severity
	return et
}

// StatusCode sets the HTTP status code associated with the error.
func (et *ErrorTemplate) StatusCode(statusCode int) *ErrorTemplate {
	et.statusCode = statusCode
	return et
}

// Protected marks the error as protected, indicating it should not be exposed externally.
func (et *ErrorTemplate) Protected(protected bool) *ErrorTemplate {
	et.protected = protected
	return et
}
