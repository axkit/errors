package errors

import (
	"encoding/json"
	se "errors"
	"strings"

	"github.com/tidwall/sjson"
)

var ErrMarshalError = Template("error marshaling failed").Severity(Critical).StatusCode(500)

type ErrorSerializationRule uint8

const (

	// AddStack - add stack in the JSON.
	AddStack ErrorSerializationRule = 1 << iota

	// AddProtected - add protected errors to the "wrapped" list in the JSON.
	// Without it, wrapped errors marked with Protected(true) are skipped.
	// The top-level error is always written as is: it should not be protected,
	// wrap it with a public template before returning it to the client.
	AddProtected

	// AddFields - add custom fields set by Set() in the JSON.
	// Fields listed in WithRootLevelFields are written regardless of this rule.
	AddFields

	// AddWrappedErrors - add previous errors in the JSON.
	AddWrappedErrors

	// IndentJSON - write indented JSON.
	IndentJSON
)

type ErrorFormattingOptions struct {
	stopStackOn     string
	include         ErrorSerializationRule
	rootLevelFields []string
}

type Option func(*ErrorFormattingOptions)

// WithStopStackOn sets the function name to stop the adding stack frames.
// The first frame whose function name contains the string and all frames after it are not written.
// As instance: WithStopStackOn("fasthttp") will stop adding stack frames
// when the function name contains "fasthttp". It's useful to avoid adding
// stack frames of the libraries which are not interesting for the user.
func WithStopStackOn(stopOnFuncContains string) Option {
	return func(e *ErrorFormattingOptions) {
		e.stopStackOn = stopOnFuncContains
	}
}

// WithAttributes sets which parts of the error are written to the output.
// Without it only message, severity, code and status code are written.
func WithAttributes(rule ErrorSerializationRule) Option {
	return func(e *ErrorFormattingOptions) {
		e.include = rule
	}
}

// WithRootLevelFields moves the listed fields from "fields" to the root of the JSON object.
// The listed fields are written even if AddFields is not set.
func WithRootLevelFields(fields []string) Option {
	return func(e *ErrorFormattingOptions) {
		e.rootLevelFields = fields
	}
}

// Predefined sets of serialization rules.
// A typical setup writes ServerOutputFormat to the server log and sends
// ClientOutputFormat to the client in production or ClientDebugOutputFormat
// in development, so the client gets the same JSON as the log.
const (
	ServerOutputFormat      = AddProtected | AddStack | AddFields | AddWrappedErrors
	ServerDebugOutputFormat = AddProtected | AddStack | AddFields | AddWrappedErrors | IndentJSON
	ClientDebugOutputFormat = AddProtected | AddStack | AddFields | AddWrappedErrors
	ClientOutputFormat      = 0 // no fields, no stack, no wrapped errors, only message.
)

// SerializedError is serialization ready error.
type SerializedError struct {
	Message    string            `json:"msg"`
	Severity   string            `json:"severity,omitempty"`
	Code       string            `json:"code,omitempty"`
	StatusCode int               `json:"statusCode,omitempty"`
	Fields     map[string]any    `json:"fields,omitempty"`
	Wrapped    []SerializedError `json:"wrapped,omitempty"`
	Stack      []StackFrame      `json:"stack,omitempty"`
}

// Serialize serializes the error to a SerializedError struct.
func Serialize(err error, opts ...Option) *SerializedError {

	if err == nil {
		return nil
	}
	var option ErrorFormattingOptions
	for _, opt := range opts {
		opt(&option)
	}
	return serialize(err, option)
}

func serialize(err error, option ErrorFormattingOptions) *SerializedError {
	switch e := err.(type) {
	case *ErrorTemplate:
		return serializeError(e.New(), option)
	case *Error:
		return serializeError(e, option)
	case interface{ Error() string }:
		return &SerializedError{
			Message: e.Error(),
		}
	}
	panic("unsupported error type")
}

func serializeError(we *Error, option ErrorFormattingOptions) *SerializedError {

	resp := SerializedError{
		Message:    we.message,
		Severity:   we.severity.String(),
		Code:       we.code,
		StatusCode: we.statusCode,
	}

	if option.include&AddFields != 0 {
		resp.Fields = cloneMap(we.fields)
	}

	if option.include&AddWrappedErrors != 0 {
		resp.Wrapped = serializeWrapped(we.err, option.include&AddProtected != 0)
	}

	if option.include&AddStack != 0 && len(we.stack) > 0 {
		resp.Stack = we.stack
		if option.stopStackOn != "" {
			for i, f := range we.stack {
				if strings.Contains(f.Function, option.stopStackOn) {
					resp.Stack = we.stack[:i]
					break
				}
			}
		}
	}
	return &resp
}

// serializeWrapped returns the chain of errors below the top-level one.
// Errors without own attributes are skipped. Protected errors are skipped
// unless addProtected is true. The chain ends at the first error that is
// not *Error or *ErrorTemplate, its full message is written as is.
func serializeWrapped(err error, addProtected bool) []SerializedError {
	var res []SerializedError

	add := func(m metadata) {
		if m.empty() || (m.protected && !addProtected) {
			return
		}
		res = append(res, SerializedError{
			Message:    m.message,
			Severity:   m.severity.String(),
			Code:       m.code,
			StatusCode: m.statusCode,
		})
	}

	for err != nil {
		switch x := err.(type) {
		case *Error:
			add(x.metadata)
			err = x.err
		case *ErrorTemplate:
			add(x.metadata)
			err = nil
		default:
			// The message of a standard wrapper, e.g. fmt.Errorf("...: %w", err),
			// includes messages of wrapped errors, so it is skipped
			// if the chain contains a protected error.
			if addProtected || !hasProtected(err) {
				res = append(res, SerializedError{Message: err.Error(), Severity: Unknown.String()})
			}
			err = nil
		}
	}
	return res
}

// hasProtected reports whether the error chain contains a protected error.
func hasProtected(err error) bool {
	var e *Error
	for err != nil {
		if se.As(err, &e) {
			if e.protected {
				return true
			}
			err = e.err
			continue
		}
		var t *ErrorTemplate
		return se.As(err, &t) && t.protected
	}
	return false
}

// ToJSON serializes the error to JSON format.
func ToJSON(err error, opts ...Option) []byte {

	if err == nil {
		return nil
	}

	var option ErrorFormattingOptions
	for _, opt := range opts {
		opt(&option)
	}

	serr := serialize(err, option)

	var rootLevelFields map[string]any
	if len(option.rootLevelFields) > 0 {
		var fields map[string]any
		switch e := err.(type) {
		case *Error:
			fields = e.fields
		case *ErrorTemplate:
			fields = e.fields
		}
		for _, k := range option.rootLevelFields {
			v, ok := fields[k]
			if !ok {
				continue
			}
			if rootLevelFields == nil {
				rootLevelFields = make(map[string]any, len(option.rootLevelFields))
			}
			rootLevelFields[k] = v
			// serr.Fields is a copy, the error itself is not modified.
			delete(serr.Fields, k)
		}
		if len(serr.Fields) == 0 {
			serr.Fields = nil
		}
	}

	var buf []byte
	var marshalErr error
	if option.include&IndentJSON != 0 {
		buf, marshalErr = json.MarshalIndent(serr, "", "  ")
	} else {
		buf, marshalErr = json.Marshal(serr)
	}

	if marshalErr == nil {
		for _, k := range option.rootLevelFields {
			if v, ok := rootLevelFields[k]; ok {
				buf, _ = sjson.SetBytes(buf, k, v)
			}
		}
		return buf
	}

	if alarmer != nil {
		alarmer.Alarm(ErrMarshalError.Wrap(marshalErr))
	}

	// Marshalling can fail if Fields contains non-serializable values.
	if len(serr.Fields) > 0 {
		serr.Fields = nil
		buf, marshalErr = json.Marshal(serr)
		if marshalErr == nil {
			return buf
		}
	}

	msg, _ := json.Marshal(marshalErr.Error())
	return []byte(`{"msg":` + string(msg) + `,"severity":"critical"}`)
}
