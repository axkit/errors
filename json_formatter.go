package errors

import (
	"encoding/json"

	"github.com/tidwall/sjson"
)

type ErrorMarshaler func(
	message string,
	severity SeverityLevel,
	statusCode int,
	code string) []byte

var errorMarshaler = DefaultErrorMarshaler

type ErrorMarshalingRule uint8

const (

	// AddStack - add stack in the JSON.
	AddStack ErrorMarshalingRule = 1 << iota

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
	include         ErrorMarshalingRule
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

func WithAttributes(rule ErrorMarshalingRule) Option {
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
	ClientJSONFormatterFlag = 0
)

// JSONErrorResponse is a struct that represents the JSON response of an error.
type JSONErrorResponse struct {
	Message    string         `json:"msg"`
	Severity   string         `json:"severity"`
	Code       string         `json:"code,omitempty"`
	StatusCode int            `json:"statusCode,omitempty"`
	Fields     map[string]any `json:"fields,omitempty"`
	Wrapped    []Error        `json:"wrapped,omitempty"`
	Stack      []StackFrame   `json:"stack,omitempty"`
}

func errorToJSON(we *Error, options ErrorFormattingOptions) ([]byte, error) {
	resp := JSONErrorResponse{
		Message:    we.message,
		Severity:   we.severity.String(),
		Code:       we.code,
		StatusCode: we.statusCode,
	}

	if options.include&AddWrappedErrors != 0 {
		resp.Wrapped = we.WrappedErrors()
		resp.Wrapped = resp.Wrapped[1:]
	}

	var roolLevelFields map[string]any

	if options.include&AddFields != 0 && len(we.fields) > 0 {
		if len(options.rootLevelFields) > 0 {
			roolLevelFields = make(map[string]any)
			resp.Fields = make(map[string]any, len(we.fields))
			for k, v := range we.fields {
				found := false
				for _, f := range options.rootLevelFields {
					if k == f {
						roolLevelFields[k] = v
						found = true
						break
					}
				}
				if !found {
					resp.Fields[k] = v
				}
			}
		} else {
			resp.Fields = we.fields
		}

		if options.include&AddStack != 0 && len(we.stack) > 0 {
			resp.Stack = we.stack
		}

	}

	var buf []byte
	var marshalError error
	if options.include&IndentJSON != 0 {
		buf, marshalError = json.MarshalIndent(resp, "", "  ")
	} else {
		buf, marshalError = json.Marshal(resp)
	}

	if marshalError != nil {
		return nil, marshalError
	}

	if len(roolLevelFields) > 0 {
		for k, v := range roolLevelFields {
			buf, _ = sjson.SetBytes(buf, k, v)
		}
	}

	return buf, nil
}

func ToJSON(err error, opts ...Option) []byte {

	var option ErrorFormattingOptions
	for _, opt := range opts {
		opt(&option)
	}

	if err == nil {
		return nil
	}

	var buf []byte
	var marshalError error
	switch e := err.(type) {
	case *Error:
		buf, marshalError = errorToJSON(e, option)
	case interface{ Error() string }:
		buf = []byte(`{"msg": "` + e.Error() + `"}`)
	}

	if marshalError != nil {
		return []byte(`{"msg": "` + err.Error() + `", "severity": "critical"}`)
	}

	return buf
}

func (err *Error) MarshalJSON() ([]byte, error) {
	return errorMarshaler(err.message, err.severity, err.statusCode, err.code), nil
}

func DefaultErrorMarshaler(message string, severity SeverityLevel, statusCode int, code string) []byte {
	buf := make([]byte, 0, len(message)+len(code)+len(severity.String())+32)
	if message != "" {
		buf, _ = sjson.SetBytes(buf, "message", message)
	}
	if severity != 0 {
		buf, _ = sjson.SetBytes(buf, "severity", severity.String())
	}
	if statusCode != 0 {
		buf, _ = sjson.SetBytes(buf, "statusCode", statusCode)
	}
	if code != "" {
		buf, _ = sjson.SetBytes(buf, "code", code)
	}
	return buf
}
