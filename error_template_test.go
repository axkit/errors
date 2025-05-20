package errors

import (
	"net/http"
	"os"
	"testing"
)

func TestTemplate(t *testing.T) {
	msg := "test error"
	err := Template(msg)
	if err.message != msg {
		t.Errorf("expected message %q, got %q", msg, err.message)
	}

	if text := err.Error(); text != msg {
		t.Errorf("expected error message %q, got %q", msg, text)
	}
}

func TestErrorTemplate_Wrap(t *testing.T) {

	var (
		filename               = "non-existing-file"
		ErrFileNotFound        = Template("file not found")
		ErrConfigReadingFailed = Template("config reading failed").
					StatusCode(http.StatusInternalServerError).
					Severity(Critical)
	)

	_, osErr := os.OpenFile(filename, os.O_RDONLY, 0)
	if osErr == nil {
		t.Fatalf("expected error opening file %q", filename)
	}
	t.Run("correct Error", func(t *testing.T) {

		noFileErr := ErrFileNotFound.Wrap(osErr).Set("file", filename)

		wErr := ErrConfigReadingFailed.Wrap(noFileErr)

		if len(wErr.stack) == 0 {
			t.Error("expected stack trace to be populated")
		}

		if len(wErr.fields) == 0 {
			t.Error("expected fields to be populated")
		}

		if fn := wErr.fields["file"]; fn != filename {
			t.Errorf("expected field %q to be %q, got %q", "file", filename, fn)
		}

		if wErr.err != noFileErr {
			t.Errorf("expected wrapped error %v, got %v", noFileErr, wErr.err)
		}

		if wErr.err.Error() != noFileErr.Error() {
			t.Errorf("expected message %q, got %q", wErr.metadata.message, osErr.Error())
		}
	})
	t.Run("manually created Error", func(t *testing.T) {
		incorrectErr := &Error{}
		err := ErrConfigReadingFailed.Wrap(incorrectErr)
		if err.stack == nil {
			t.Error("expected stack trace to be populated")
		}
	})
}

func TestErrorTemplate_New(t *testing.T) {
	et := Template("predefined error")
	err := et.New()

	if err.metadata.message != et.metadata.message {
		t.Errorf("expected message %q, got %q", et.metadata.message, err.metadata.message)
	}

	if len(err.stack) == 0 {
		t.Error("expected stack trace to be populated")
	}
}
func TestErrorTemplate_Error(t *testing.T) {
	msg := "predefined error"
	et := Template(msg)

	if et.Error() != msg {
		t.Errorf("expected error message %q, got %q", msg, et.Error())
	}
}

func TestErrorTemplate_Set(t *testing.T) {

	et := Template("predefined error")

	t.Run("Set", func(t *testing.T) {
		key, value := "key", "value"
		et.Set(key, value)
		if et.fields[key] != value {
			t.Errorf("expected field %q to be %q, got %q", key, value, et.fields[key])
		}
	})

	t.Run("Code", func(t *testing.T) {
		code := "E123"
		et.Code(code)
		if et.code != code {
			t.Errorf("expected code %q, got %q", code, et.code)
		}
	})
	t.Run("Severity", func(t *testing.T) {
		severity := SeverityLevel(Medium)
		et.Severity(severity)
		if et.severity != severity {
			t.Errorf("expected severity %v, got %v", severity, et.severity)
		}
	})
	t.Run("StatusCode", func(t *testing.T) {
		statusCode := http.StatusBadRequest
		et.StatusCode(statusCode)
		if et.statusCode != statusCode {
			t.Errorf("expected status code %d, got %d", statusCode, et.statusCode)
		}
	})
	t.Run("Protected", func(t *testing.T) {
		protected := true
		et.Protected(protected)
		if et.protected != protected {
			t.Errorf("expected protected %v, got %v", protected, et.protected)
		}
	})
}

func TestCloneMap(t *testing.T) {
	pe := Template("predefined error")
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
