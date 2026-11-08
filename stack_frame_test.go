package errors

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"testing"
)

const pkgPrefix = "github.com/axkit/errors."

func TestDefaultCallerFrames(t *testing.T) {
	t.Run("offset 0 starts at runtime.Callers", func(t *testing.T) {
		frames := DefaultCallerFrames(0)
		if len(frames) == 0 || frames[0].Function != "runtime.Callers" {
			t.Fatalf("expected first frame runtime.Callers, got %v", frames)
		}
	})

	t.Run("offset 2 starts at the caller", func(t *testing.T) {
		frames := DefaultCallerFrames(2)
		if len(frames) == 0 {
			t.Fatal("expected frames, got none")
		}
		want := pkgPrefix + "TestDefaultCallerFrames.func2"
		if frames[0].Function != want {
			t.Errorf("expected first frame %q, got %q", want, frames[0].Function)
		}
	})

	t.Run("frame content", func(t *testing.T) {
		for _, f := range DefaultCallerFrames(2) {
			if f.Function == "" || f.Line == 0 {
				t.Errorf("incomplete frame: %+v", f)
			}
			if !strings.HasSuffix(f.File, ":"+strconv.Itoa(f.Line)) {
				t.Errorf("expected file to end with line number, got %q", f.File)
			}
			if f.Function == "runtime.goexit" {
				t.Error("runtime.goexit must not be included")
			}
		}
	})

	t.Run("large offset", func(t *testing.T) {
		if frames := DefaultCallerFrames(100); len(frames) != 0 {
			t.Errorf("expected no frames, got %v", frames)
		}
	})
}

func TestDefaultCallerFramesMaxLen(t *testing.T) {
	var frames []StackFrame
	var deep func(n int)
	deep = func(n int) {
		if n == 0 {
			frames = DefaultCallerFrames(0)
			return
		}
		deep(n - 1)
	}
	deep(CallingStackMaxLen * 2)

	if len(frames) != CallingStackMaxLen {
		t.Errorf("expected %d stack frames, got %d", CallingStackMaxLen, len(frames))
	}
}

func TestStackFrameString(t *testing.T) {
	tests := []struct {
		name  string
		frame StackFrame
		want  string
	}{
		{"file with line", StackFrame{Function: "main.f", File: "/a/b.go:42", Line: 42}, "/a/b.go:42 main.f"},
		{"file without line", StackFrame{Function: "main.f", File: "/a/b.go", Line: 42}, "/a/b.go:42 main.f"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.frame.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestStackStartsAtCaller checks that every way of creating an error
// records the stack starting at the line that created it.
func TestStackStartsAtCaller(t *testing.T) {
	tpl := Template("tpl")
	want := pkgPrefix + "TestStackStartsAtCaller"

	tests := []struct {
		name string
		err  *Error
	}{
		{"ErrorTemplate.New", tpl.New()},
		{"ErrorTemplate.Wrap std error", tpl.Wrap(io.EOF)},
		{"ErrorTemplate.Wrap template", tpl.Wrap(Template("inner"))},
		{"ErrorTemplate.Wrap Error without stack", tpl.Wrap(&Error{})},
		{"Error.Wrap std error", (&Error{}).Wrap(io.EOF)},
		{"Error.Wrap template", (&Error{}).Wrap(Template("inner"))},
		{"Error.Wrap Error without stack", (&Error{}).Wrap(&Error{})},
		{"Wrap", Wrap(io.EOF, "msg")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.err.stack) == 0 {
				t.Fatal("expected stack, got none")
			}
			if got := tt.err.stack[0].Function; got != want {
				t.Errorf("expected first frame %q, got %q", want, got)
			}
		})
	}
}

func TestWrapKeepsOriginalStack(t *testing.T) {
	inner := Template("inner").New()
	outer := Template("outer").Wrap(inner)
	if &outer.stack[0] != &inner.stack[0] {
		t.Error("expected wrapped error to keep the original stack")
	}
}

func TestCallerFramesFuncOverride(t *testing.T) {
	saved := CallerFramesFunc
	defer func() { CallerFramesFunc = saved }()

	var offsets []int
	CallerFramesFunc = func(offset int) []StackFrame {
		offsets = append(offsets, offset)
		return []StackFrame{{Function: "custom"}}
	}

	tpl := Template("tpl")
	errs := []*Error{
		tpl.New(),
		tpl.Wrap(io.EOF),
		(&Error{}).Wrap(io.EOF),
		Wrap(io.EOF, "msg"),
	}

	for i, e := range errs {
		if len(e.stack) != 1 || e.stack[0].Function != "custom" {
			t.Errorf("error %d: expected stack from custom collector, got %v", i, e.stack)
		}
	}
	for _, o := range offsets {
		if o != callerFramesSkip {
			t.Errorf("expected offset %d, got %d", callerFramesSkip, o)
		}
	}
}

func TestStackFrameJSON(t *testing.T) {
	t.Run("marshal has no line key", func(t *testing.T) {
		buf, err := json.Marshal(StackFrame{Function: "main.f", File: "/a/b.go:42", Line: 42})
		if err != nil {
			t.Fatal(err)
		}
		want := `{"func":"main.f","file":"/a/b.go:42"}`
		if string(buf) != want {
			t.Errorf("got %s, want %s", buf, want)
		}
	})

	tests := []struct {
		name string
		in   string
		want StackFrame
	}{
		{"line from file", `{"func":"main.f","file":"/a/b.go:42"}`, StackFrame{"main.f", "/a/b.go:42", 42}},
		{"old format with line key", `{"func":"main.f","file":"/a/b.go:42","line":42}`, StackFrame{"main.f", "/a/b.go:42", 42}},
		{"file without line", `{"func":"main.f","file":"/a/b.go"}`, StackFrame{"main.f", "/a/b.go", 0}},
		{"windows path", `{"func":"main.f","file":"C:/a/b.go:7"}`, StackFrame{"main.f", "C:/a/b.go:7", 7}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got StackFrame
			if err := json.Unmarshal([]byte(tt.in), &got); err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}

	t.Run("invalid json", func(t *testing.T) {
		var got StackFrame
		if err := json.Unmarshal([]byte(`{"func":1}`), &got); err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestToJSONStackRoundTrip(t *testing.T) {
	src := Template("tpl").New()
	var got SerializedError
	if err := json.Unmarshal(ToJSON(src, WithAttributes(AddStack)), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Stack) == 0 || got.Stack[0] != src.stack[0] {
		t.Errorf("expected first frame %+v, got %v", src.stack[0], got.Stack)
	}
}
