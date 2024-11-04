package errors

import (
	"net/http"
	"os"
	"testing"
)

func TestPredefinedError(t *testing.T) {
	msg := "test error"
	err := New(msg)
	if err.message != msg {
		t.Errorf("expected message %q, got %q", msg, err.message)
	}

	if text := err.Error(); text != msg {
		t.Errorf("expected error message %q, got %q", msg, text)
	}
}

func TestPredefinedError_Wrap(t *testing.T) {

	filename := "non-existing-file"

	var ErrConfigFileNotFound = New("config file not found").
		StatusCode(http.StatusInternalServerError).
		Severity(Critical)

	if _, osErr := os.OpenFile(filename, os.O_RDONLY, 0); osErr != nil {

		wErr := ErrConfigFileNotFound.Wrap(osErr).Set("file", filename)

		if len(wErr.stack) == 0 {
			t.Error("expected stack trace to be populated")
		}

		if len(wErr.fields) == 0 {
			t.Error("expected fields to be populated")
		}

		if fn := wErr.fields["file"]; fn != filename {
			t.Errorf("expected field %q to be %q, got %q", "file", filename, fn)
		}

		if wErr.err != osErr {
			t.Errorf("expected wrapped error %v, got %v", osErr, wErr.err)
		}

		if wErr.err.Error() != osErr.Error() {
			t.Errorf("expected message %q, got %q", wErr.attrs.message, osErr.Error())
		}
	}
}

func TestPredefinedError_Raise(t *testing.T) {
	pe := New("predefined error")
	raisedErr := pe.Raise()

	if raisedErr.attrs.message != pe.attrs.message {
		t.Errorf("expected message %q, got %q", pe.attrs.message, raisedErr.attrs.message)
	}
}

func TestPredefinedError_Set(t *testing.T) {
	pe := New("predefined error")
	key, value := "key", "value"
	pe.Set(key, value)

	if pe.fields[key] != value {
		t.Errorf("expected field %q to be %q, got %q", key, value, pe.fields[key])
	}
}

func TestPredefinedError_Attrs(t *testing.T) {
	pe := New("predefined error")

	// Test Code
	code := "E123"
	pe.Code(code)
	if pe.code != code {
		t.Errorf("expected code %q, got %q", code, pe.code)
	}

	// Test Severity
	severity := SeverityLevel(1)
	pe.Severity(severity)
	if pe.severity != severity {
		t.Errorf("expected severity %v, got %v", severity, pe.severity)
	}

	// Test StatusCode
	statusCode := http.StatusBadRequest
	pe.StatusCode(statusCode)
	if pe.statusCode != statusCode {
		t.Errorf("expected status code %d, got %d", statusCode, pe.statusCode)
	}

	// Test Protected
	protected := true
	pe.Protected(protected)
	if pe.protected != protected {
		t.Errorf("expected protected %v, got %v", protected, pe.protected)
	}
}

func TestPredefinedError_CloneFields(t *testing.T) {
	pe := New("predefined error")
	pe.Set("key1", "value1")
	pe.Set("key2", "value2")

	clonedFields := cloneMap(pe.fields)
	if len(clonedFields) != len(pe.fields) {
		t.Errorf("expected cloned fields length %d, got %d", len(pe.fields), len(clonedFields))
	}
	for k, v := range pe.fields {
		if clonedFields[k] != v {
			t.Errorf("expected field %q to be %q, got %q", k, v, clonedFields[k])
		}
	}
}
