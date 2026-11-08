package errors

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// StackFrame describes content of a single stack frame stored with error.
//
// File holds "path:line", so IDEs and terminals open it on click.
// Line holds the same line number for use in Go code. It is not written
// to JSON to avoid duplication and is restored from File on unmarshal.
type StackFrame struct {
	Function string `json:"func"`
	File     string `json:"file"`
	Line     int    `json:"-"`
}

// UnmarshalJSON implements json.Unmarshaler. Line is taken from the
// "line" key if present (JSON written before v1.1.0), otherwise
// it is parsed from the ":line" suffix of File.
func (s *StackFrame) UnmarshalJSON(data []byte) error {
	var v struct {
		Function string `json:"func"`
		File     string `json:"file"`
		Line     int    `json:"line"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	s.Function, s.File, s.Line = v.Function, v.File, v.Line
	if s.Line == 0 {
		if i := strings.LastIndexByte(s.File, ':'); i >= 0 {
			if n, err := strconv.Atoi(s.File[i+1:]); err == nil {
				s.Line = n
			}
		}
	}
	return nil
}

// String returns the frame as "file:line function".
// File collected by DefaultCallerFrames already ends with ":line",
// in that case the line is not appended twice.
func (s StackFrame) String() string {
	file := s.File
	if !strings.HasSuffix(file, ":"+strconv.Itoa(s.Line)) {
		file += ":" + strconv.Itoa(s.Line)
	}
	return fmt.Sprintf("%s %s", file, s.Function)
}

// callerFramesSkip is the offset passed to CallerFramesFunc by functions of this
// package, so the first collected frame is the code that created or wrapped the error.
// It skips runtime.Callers, the collector itself and the package function.
const callerFramesSkip = 3

var (
	// CallerFramesFunc holds the function used to collect call frames
	// when an error is created or wrapped. The offset has the same meaning
	// as the skip argument of runtime.Callers called inside the collector.
	CallerFramesFunc func(offset int) []StackFrame = DefaultCallerFrames

	// CallingStackMaxLen holds maximum elements in the call frames.
	CallingStackMaxLen int = 16
)

// DefaultCallerFrames returns default implementation of call frames collector.
// The offset is the number of frames to skip, as in runtime.Callers:
// 0 is runtime.Callers itself, 1 is DefaultCallerFrames, 2 is its caller.
// The runtime.goexit frame is never included.
func DefaultCallerFrames(offset int) []StackFrame {
	pc := make([]uintptr, CallingStackMaxLen)
	n := runtime.Callers(offset, pc)
	if n == 0 {
		return nil
	}

	res := make([]StackFrame, 0, n)
	frames := runtime.CallersFrames(pc[:n])
	for {
		frame, more := frames.Next()
		if frame.Function != "runtime.goexit" {
			res = append(res, StackFrame{
				Function: frame.Function,
				File:     frame.File + ":" + strconv.Itoa(frame.Line),
				Line:     frame.Line,
			})
		}
		if !more {
			break
		}
	}
	return res
}
