package errors

import (
	"fmt"
	"testing"
)

func TestDefaultCallerFrames(t *testing.T) {
	tests := []struct {
		name           string
		offset         int
		expectedFrames int
	}{
		{"NoOffset", 0, 1},
		{"WithOffset", 2, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frames := DefaultCallerFrames(tt.offset)
			if len(frames) != tt.expectedFrames {
				t.Errorf("Expected %d stack frames, got %d", tt.expectedFrames, len(frames))
			}

			fmt.Printf("frames: %v\n", frames)

			for _, frame := range frames {
				if frame.Function == "" {
					t.Errorf("Expected function name, got empty string")
				}
				if frame.File == "" {
					t.Errorf("Expected file name, got empty string")
				}
				if frame.Line == 0 {
					t.Errorf("Expected line number, got 0")
				}
			}
		})
	}
}

func TestDefaultCallerFramesMaxLen(t *testing.T) {
	frames := DefaultCallerFrames(0)
	if len(frames) > CallingStackMaxLen {
		t.Errorf("Expected maximum %d stack frames, got %d", CallingStackMaxLen, len(frames))
	}
}

func TestDefaultCallerFramesContent(t *testing.T) {
	// This test ensures that the DefaultCallerFrames function captures the correct frame
	frames := DefaultCallerFrames(0)
	if len(frames) == 0 {
		t.Fatalf("Expected non-empty stack frames, got empty")
	}

	// Check the first frame to ensure it matches the test function
	frame := frames[0]
	if frame.Function != "testing.tRunner" {
		t.Errorf("Expected function 'runtime.Callers', got %s", frame.Function)
	}
	if frame.File == "" {
		t.Errorf("Expected file name, got empty string")
	}
	if frame.Line == 0 {
		t.Errorf("Expected line number, got 0")
	}
}
