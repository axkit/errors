package errors

import (
	"encoding/json"
	"slices"

	"github.com/tidwall/sjson"
)

var ErrMarshalError = New("error marshaling failed").Severity(Critical).StatusCode(500)

type ErrorSerializationRule uint8

const (

	// AddStack - add stack in the JSON.
	AddStack ErrorSerializationRule = 1 << iota

	// AddProtected - add protected errors in the JSON.
	AddProtected

	// AddFields - add fields in the JSON.
	AddFields

	// AddWrappedErrors - add previous errors in the JSON.
	AddWrappedErrors

	IndentJSON
)

type ErrorFormattingOptions struct {
	stopStackOn     string
	include         ErrorSerializationRule
	rootLevelFields []string
}

type Option func(*ErrorFormattingOptions)

// WithStopStackOn sets the function name to stop the adding stack frames.
// As instance: WithStopStackOn("fasthttp") will stop adding stack frames
// when the function name contains "fasthttp". It's useful to avoid adding
// stack frames of the libraries which are not interesting for the user.
func WithStopStackOn(stopOnFuncContains string) Option {
	return func(e *ErrorFormattingOptions) {
		e.stopStackOn = stopOnFuncContains
	}
}

func WithAttributes(rule ErrorSerializationRule) Option {
	return func(e *ErrorFormattingOptions) {
		e.include = rule
	}
}

func WithRootLevelFields(fields []string) Option {
	return func(e *ErrorFormattingOptions) {
		e.rootLevelFields = fields
	}
}

const (
	ServerOutputFormat      = AddProtected | AddStack | AddFields | AddWrappedErrors
	ServerDebugOutputFormat = AddProtected | AddStack | AddFields | AddWrappedErrors | IndentJSON
	ClientDebugOutputFormat = AddProtected | AddStack | AddFields | AddWrappedErrors
	ClientOutputFormat      = 0 // no fields, no stack, no wrapped errors, only message.
)

// SerializedError is serialization ready error.
type SerializedError struct {
	Message    string         `json:"msg"`
	Severity   string         `json:"severity,omitempty"`
	Code       string         `json:"code,omitempty"`
	StatusCode int            `json:"statusCode,omitempty"`
	Fields     map[string]any `json:"fields,omitempty"`
	Wrapped    []Error        `json:"wrapped,omitempty"`
	Stack      []StackFrame   `json:"stack,omitempty"`
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
		return serializeError(e.Build(), option)
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
		Fields:     we.fields,
		Wrapped:    nil,
		Stack:      nil,
	}

	if option.include&AddWrappedErrors != 0 {
		resp.Wrapped = we.WrappedErrors()
		resp.Wrapped = resp.Wrapped[1:]
	}

	if option.include&AddStack != 0 && len(we.stack) > 0 {
		resp.Stack = we.stack
	}
	return &resp
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
		rootLevelFields = make(map[string]any)
		for k, v := range serr.Fields {
			if slices.Contains(option.rootLevelFields, k) {
				rootLevelFields[k] = v
				delete(serr.Fields, k)
			}
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
		if len(rootLevelFields) > 0 {
			for k, v := range rootLevelFields {
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

	return []byte(`{"msg":"` + marshalErr.Error() + `", "severity": "critical"}`)
}
