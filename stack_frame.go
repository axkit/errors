package errors

import "runtime"

// StackFrame describes content of a single stack frame stored with error.
type StackFrame struct {
	Function string `json:"func"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

var (
	// CallerFramesFunc holds default function used by function Catch()
	// to collect call frames.
	CallerFramesFunc func(offset int) []StackFrame = DefaultCallerFrames

	// CallingStackMaxLen holds maximum elements in the call frames.
	CallingStackMaxLen int = 15
)

// DefaultCallerFrames returns default implementation of call frames collector.
func DefaultCallerFrames(offset int) []StackFrame {
	res := make([]StackFrame, 0, 12)
	pc := make([]uintptr, CallingStackMaxLen)
	n := runtime.Callers(3+offset, pc)
	frames := runtime.CallersFrames(pc[:n])

	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		res = append(res, StackFrame{Function: frame.Function, File: frame.File, Line: frame.Line})
	}
	return res
}
