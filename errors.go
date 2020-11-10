package errors

import (
	"runtime"
)

// CaptureStackStopWord captures stack till the function name
// has not contains CaptureStackStopWord
//
// The stack capturing ignored if it's empty or it's appeared
// in the first function name.
var CaptureStackStopWord string

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

// MarshalJSON implements json/Marsaller interface.
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

// Frame describes content of a single stack frame stored with error.
type Frame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

var (
	// DefaultCallerFramesFunc holds default function used by function Catch()
	// to collect call frames.
	DefaultCallerFramesFunc func(offset int) []Frame = CallerFrames

	// CallingStackMaxLen holds maximum elements in the call frames.
	CallingStackMaxLen int = 15
)

// CallerFrames returns not more then slice of Frame.
func CallerFrames(offset int) []Frame {
	var res []Frame
	pc := make([]uintptr, CallingStackMaxLen)
	n := runtime.Callers(3+offset, pc)
	frames := runtime.CallersFrames(pc[:n])

	for {
		frame, more := frames.Next()
		res = append(res, Frame{Function: frame.Function, File: frame.File, Line: frame.Line})
		if !more {
			break
		}
	}
	return res
}

// NotFound объект не найден.
var NotFound = func(msg string) *CatchedError {
	return newx(msg).StatusCode(404).Severity(Medium)
}

var ValidationFailed = func(msg string) *CatchedError {
	return newx(msg).StatusCode(400).Severity(Tiny)
}

var ConsistencyFailed = func() *CatchedError {
	return newx("consistency failed").StatusCode(500).Severity(Critical)
}

var InvalidRequestBody = func(s string) *CatchedError {
	return newx(s).StatusCode(400).Severity(Critical)
}

var Unauthorized = func() *CatchedError {
	return newx("unauthorized").StatusCode(401).Severity(Medium)
}

var Forbidden = func() *CatchedError {
	return newx("forbidden").StatusCode(403).Severity(Critical)
}

var InternalError = func() *CatchedError {
	return newx("internal error").StatusCode(500).Severity(Critical)
}

var UnprocessableEntity = func(s string) *CatchedError {
	return newx(s).StatusCode(422).Severity(Medium)
}
