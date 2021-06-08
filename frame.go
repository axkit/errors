package errors

import "runtime"

// Frame describes content of a single stack frame stored with error.
type Frame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

var (
	// CallerFramesFunc holds default function used by function Catch()
	// to collect call frames.
	CallerFramesFunc func(offset int) []Frame = DefaultCallerFrames

	// CallingStackMaxLen holds maximum elements in the call frames.
	CallingStackMaxLen int = 15
)

// DefaultCallerFrames returns default implementation of call frames collector.
func DefaultCallerFrames(offset int) []Frame {
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
