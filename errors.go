package errors

import (
	e "errors"
)

// CaptureStackStopWord holds a function name part. Stack will be captured
// till the function with the first CaptureStackStopWord
//
// The stack capturing ignored if it's empty or it's appeared
// in the first function name.
//
// It can be used to ignore stack above HTTP handler.
var CaptureStackStopWord string = "fasthttp"

// RootLevelFields holds a name of context fields which will be generated
// placed on the root level in JSON together with standard error attribute
// such as msg, code, statusCode. All other context fields will be located
// under root level attribute "ctx".
var RootLevelFields = []string{"reason"}

// Mode describes allowed methods of response returned Error().
type Mode int

const (
	// Multi return all error messages separated by ": ".
	Multi Mode = 0

	// Single return message of last error in the stack.
	Single Mode = 1
)

// ErrorMethodMode holds behavior of Error() method.
var ErrorMethodMode = Single

// SeverityLevel describes error severity levels.
type SeverityLevel int

const (
	// Tiny classifies as expected, managed errors that do not require administrator attention.
	// It's not recommended to write a call stack to the journal file.
	//
	// Example: error related to validation of entered form fields.
	Tiny SeverityLevel = iota

	// Medium classifies an regular error. A call stack is written to the log.
	Medium

	// Critical classifies a significant error, requiring immediate attention.
	// An error occurrence fact shall be passed to the administrator in all possible ways.
	// A call stack is written to the log.
	Critical
)

var (
	// important: quotas are included.
	tiny     = []byte(`"tiny"`)
	medium   = []byte(`"medium"`)
	critical = []byte(`"critical"`)
	unknown  = []byte(`"unknown"`)
)

const (
	stiny     = "tiny"
	smedium   = "medium"
	scritical = "critical"
	sunknown  = "unknown"
)

// String returns severity level string representation.
func (sl SeverityLevel) String() string {
	switch sl {
	case Tiny:
		return stiny
	case Medium:
		return smedium
	case Critical:
		return scritical
	}
	return sunknown
}

// MarshalJSON implements json/Marshaller interface.
func (sl SeverityLevel) MarshalJSON() ([]byte, error) {
	switch sl {
	case Tiny:
		return tiny, nil
	case Medium:
		return medium, nil
	case Critical:
		return critical, nil
	}
	return unknown, nil
}

// TODO: delete all predefined errors!
var (
	nf  = newx("not found", false).StatusCode(404).Severity(Medium)
	vf  = newx("validation failed", false).StatusCode(400)
	cf  = newx("consistency failed", false).StatusCode(500).Severity(Critical)
	irr = newx("invalid request body", false).StatusCode(400).Severity(Critical)
	ie  = newx("internal error", false).StatusCode(500).Severity(Critical)
	un  = newx("unauthorized", false).StatusCode(401).Severity(Medium)
	fbd = newx("forbidden", false).StatusCode(403).Severity(Critical)
	ue  = newx("unprocessable entity", false).StatusCode(422).Severity(Medium)
)

// NotFound is a function, returns *CatchedError with predefined StatusCode=404 and Severity=Medium.

var NotFound = func(msg string) *CatchedError {
	return nf.Raise().msg(msg)
}

// ValidationFailed is a function, returns *CatchedError with predefined StatusCode=400 and Severity=Tiny.
var ValidationFailed = func(msg string) *CatchedError {
	return vf.Raise().msg(msg)
}

// ConsistencyFailed is a function, returns *CatchedError with predefined StatusCode=500 and Severity=Critical.
var ConsistencyFailed = func() *CatchedError {
	return cf.Raise()
}

var InvalidRequestBody = func(s string) *CatchedError {
	return irr.Raise().msg(s)
}

// Unauthorized is a function, returns *CatchedError with predefined StatusCode=401 and Severity=Medium.
var Unauthorized = func() *CatchedError {
	return ue.Raise()
}

// Forbidden is a function, returns *CatchedError with predefined StatusCode=403 and Severity=Critical.
var Forbidden = func() *CatchedError {
	return fbd.Raise()
}

// ValidationFailed is a function, returns *CatchedError with predefined StatusCode=400 and Severity=Tiny.
var InternalError = func() *CatchedError {
	return ie.Raise()
}

var UnprocessableEntity = func(s string) *CatchedError {
	return ue.Raise().msg(s)
}

func Is(err, target error) bool {
	ce, cok := err.(*CatchedError)
	te, tok := target.(*CatchedError)
	if cok && tok {
		return ce.uid == te.uid
	}
	return e.Is(err, target)
}

func As(err error, target interface{}) bool {
	return e.As(err, target)
}

// Unwrap returns the result of calling the Unwrap method on err, if err's type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
func Unwrap(err error) error {
	return e.Unwrap(err)
}
