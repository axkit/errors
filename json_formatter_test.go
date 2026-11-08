package errors

import (
	"encoding/json"
	se "errors"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestWithStopStackOn(t *testing.T) {
	opt := WithStopStackOn("testing")
	options := &ErrorFormattingOptions{}
	opt(options)
	if options.stopStackOn != "testing" {
		t.Errorf("expected 'fasthttp', got '%s'", options.stopStackOn)
	}
}

func TestWithAttributes(t *testing.T) {
	opt := WithAttributes(AddStack | AddFields)
	options := &ErrorFormattingOptions{}
	opt(options)
	if options.include != (AddStack | AddFields) {
		t.Errorf("expected '%v', got '%v'", AddStack|AddFields, options.include)
	}
}

func TestWithRootLevelFields(t *testing.T) {
	fields := []string{"field1", "field2"}
	opt := WithRootLevelFields(fields)
	options := &ErrorFormattingOptions{}
	opt(options)
	if !reflect.DeepEqual(options.rootLevelFields, fields) {
		t.Errorf("expected '%v', got '%v'", fields, options.rootLevelFields)
	}
}

func TestToJSON(t *testing.T) {

	errorTemplate := Template("test error").Severity(Tiny).Code("E500").StatusCode(500)
	err := errorTemplate.Wrap(io.EOF)

	options := []Option{
		WithAttributes(AddStack | AddFields | AddWrappedErrors),
		WithRootLevelFields([]string{"field1"}),
	}

	jsonBytes := ToJSON(err, options...)
	var jsonResponse SerializedError
	unmarshalErr := json.Unmarshal(jsonBytes, &jsonResponse)
	if unmarshalErr != nil {
		t.Errorf("unexpected error: %v", unmarshalErr)
	}
	if jsonResponse.Message != "test error" {
		t.Errorf("expected 'test error', got '%s'", jsonResponse.Message)
	}
	if jsonResponse.Severity != stiny {
		t.Errorf("expected 'tiny', got '%s'", jsonResponse.Severity)
	}
	if jsonResponse.Code != "E500" {
		t.Errorf("expected 'E500', got '%s'", jsonResponse.Code)
	}
	if jsonResponse.StatusCode != 500 {
		t.Errorf("expected 500, got %d", jsonResponse.StatusCode)
	}
	if len(jsonResponse.Wrapped) != 1 {
		t.Fatalf("expected 1 wrapped error, got %d", len(jsonResponse.Wrapped))
	}
	if jsonResponse.Wrapped[0].Message != "EOF" {
		t.Errorf("expected 'EOF', got '%s'", jsonResponse.Wrapped[0].Message)
	}
}

func TestWithStopStackOnEmpty(t *testing.T) {
	opt := WithStopStackOn("")
	options := &ErrorFormattingOptions{}
	opt(options)
	if options.stopStackOn != "" {
		t.Errorf("expected empty string, got '%s'", options.stopStackOn)
	}
}

func TestWithAttributesNone(t *testing.T) {
	opt := WithAttributes(0)
	options := &ErrorFormattingOptions{}
	opt(options)
	if options.include != 0 {
		t.Errorf("expected 0, got '%v'", options.include)
	}
}

func TestWithRootLevelFieldsEmpty(t *testing.T) {
	fields := []string{}
	opt := WithRootLevelFields(fields)
	options := &ErrorFormattingOptions{}
	opt(options)
	if len(options.rootLevelFields) != 0 {
		t.Errorf("expected empty slice, got '%v'", options.rootLevelFields)
	}
}

func TestToJSONWithContext_Error(t *testing.T) {
	innerErrorTemplate := Template("embedded error").Code("E1234").StatusCode(400)
	innerError := innerErrorTemplate.New()

	outerErrorTempate := Template("test error").Severity(Medium).Code("E500").StatusCode(500)
	err := outerErrorTempate.Wrap(innerError)

	tcases := []struct {
		options []Option
		fields  map[string]any
	}{
		{
			options: []Option{
				WithAttributes(AddStack | AddFields | AddWrappedErrors),
				WithRootLevelFields([]string{"field1"}),
			},
			fields: map[string]any{
				"field2": "value2",
			},
		},
		{
			options: []Option{
				WithAttributes(AddStack | AddFields | AddWrappedErrors),
			},
			fields: map[string]any{
				"field1": "value1",
				"field2": "value2",
			},
		},
	}

	for _, tcase := range tcases {
		err.Set("field1", "value1").Set("field2", "value2")
		jsonBytes := ToJSON(err, tcase.options...)
		var jsonResponse SerializedError
		unmarshalErr := json.Unmarshal(jsonBytes, &jsonResponse)
		if unmarshalErr != nil {
			t.Errorf("unexpected error: %v", unmarshalErr)
		}
		if jsonResponse.Message != "test error" {
			t.Errorf("expected 'test error', got '%s'", jsonResponse.Message)
		}
		if jsonResponse.Severity != smedium {
			t.Errorf("expected 'medium', got '%s'", jsonResponse.Severity)
		}
		if jsonResponse.Code != "E500" {
			t.Errorf("expected 'E500', got '%s'", jsonResponse.Code)
		}
		if jsonResponse.StatusCode != 500 {
			t.Errorf("expected 500, got %d", jsonResponse.StatusCode)
		}

		if len(jsonResponse.Fields) != len(tcase.fields) {

			t.Errorf("expected %v fields, got %v", tcase.fields, jsonResponse.Fields)
		}

		for k, v := range tcase.fields {
			if jsonResponse.Fields[k] != v {
				t.Errorf("expected '%v', got '%v'", v, jsonResponse.Fields[k])
			}
		}

		if len(jsonResponse.Stack) == 0 || jsonResponse.Stack[0].Function != pkgPrefix+"TestToJSONWithContext_Error" {
			t.Errorf("expected stack to start at TestToJSONWithContext_Error, got %v", jsonResponse.Stack)
		}

		if len(jsonResponse.Wrapped) != 1 {
			t.Errorf("expected 1 wrapped error, got %d", len(jsonResponse.Wrapped))
		}
	}
}

func TestToJSONWithContext_error(t *testing.T) {

	innerErrorTemplate := Template("embedded error").Code("E1234").StatusCode(400)
	innerError := innerErrorTemplate.Wrap(io.EOF)

	outerErrorTempate := Template("test error").Severity(Medium).Code("E500").StatusCode(500)
	err := outerErrorTempate.Wrap(innerError)

	tcases := []struct {
		options []Option
		fields  map[string]any
	}{
		{
			options: []Option{
				WithAttributes(AddStack | AddFields | AddWrappedErrors),
				WithRootLevelFields([]string{"field1"}),
			},
			fields: map[string]any{
				"field2": "value2",
			},
		},
		{
			options: []Option{
				WithAttributes(AddStack | AddFields | AddWrappedErrors),
			},
			fields: map[string]any{
				"field1": "value1",
				"field2": "value2",
			},
		},
	}

	for _, tcase := range tcases {
		err.Set("field1", "value1").Set("field2", "value2")
		jsonBytes := ToJSON(err, tcase.options...)
		var jsonResponse SerializedError
		unmarshalErr := json.Unmarshal(jsonBytes, &jsonResponse)
		if unmarshalErr != nil {
			t.Errorf("unexpected error: %v", unmarshalErr)
		}
		if jsonResponse.Message != "test error" {
			t.Errorf("expected 'test error', got '%s'", jsonResponse.Message)
		}
		if jsonResponse.Severity != smedium {
			t.Errorf("expected 'medium', got '%s'", jsonResponse.Severity)
		}
		if jsonResponse.Code != "E500" {
			t.Errorf("expected 'E500', got '%s'", jsonResponse.Code)
		}
		if jsonResponse.StatusCode != 500 {
			t.Errorf("expected 500, got %d", jsonResponse.StatusCode)
		}

		if len(jsonResponse.Fields) != len(tcase.fields) {

			t.Errorf("expected %v fields, got %v", tcase.fields, jsonResponse.Fields)
		}

		for k, v := range tcase.fields {
			if jsonResponse.Fields[k] != v {
				t.Errorf("expected '%v', got '%v'", v, jsonResponse.Fields[k])
			}
		}

		if len(jsonResponse.Stack) == 0 || jsonResponse.Stack[0].Function != pkgPrefix+"TestToJSONWithContext_error" {
			t.Errorf("expected stack to start at TestToJSONWithContext_error, got %v", jsonResponse.Stack)
		}

		if len(jsonResponse.Wrapped) != 2 {
			t.Errorf("expected 2 wrapped errors, got %d", len(jsonResponse.Wrapped))
		}
	}
}

func TestToJSON_error(t *testing.T) {

	if ToJSON(nil) != nil {
		t.Errorf("expected nil, got %v", ToJSON(nil))
	}

	err := se.New("test error")
	expectedResponse := `{"msg":"test error"}`

	if response := string(ToJSON(err)); response != expectedResponse {
		t.Errorf("expected '%s', got '%s'", expectedResponse, response)
	}
}

func TestToJSON_identMarshal(t *testing.T) {
	err := Template("test error").New()
	expectedResponse := `{
  "msg": "test error",
  "severity": "unknown"
}`

	if response := string(ToJSON(err, WithAttributes(IndentJSON))); response != expectedResponse {
		t.Errorf("expected '%s', got '%s'", expectedResponse, response)
	}
}

func jsonString(err error, rules ErrorSerializationRule, opts ...Option) string {
	return string(ToJSON(err, append([]Option{WithAttributes(rules)}, opts...)...))
}

func TestToJSONWrappedChain(t *testing.T) {
	errNotFound := Template("record not found").Code("DB-404")
	errUserLoad := Template("failed to load user").Code("USR-500")

	t.Run("all levels without top-level duplicate", func(t *testing.T) {
		err := Wrap(errUserLoad.Wrap(errNotFound.Wrap(io.EOF)), "handle request")
		want := `{"msg":"handle request","severity":"unknown","code":"USR-500","wrapped":[` +
			`{"msg":"failed to load user","severity":"unknown","code":"USR-500"},` +
			`{"msg":"record not found","severity":"unknown","code":"DB-404"},` +
			`{"msg":"EOF","severity":"unknown"}]}`
		if got := jsonString(err, AddWrappedErrors); got != want {
			t.Errorf("got  %s\nwant %s", got, want)
		}
	})

	t.Run("package Wrap three times", func(t *testing.T) {
		err := Wrap(Wrap(Wrap(io.EOF, "read header"), "parse config"), "start app")
		want := `{"msg":"start app","severity":"unknown","wrapped":[` +
			`{"msg":"parse config","severity":"unknown"},` +
			`{"msg":"read header","severity":"unknown"},` +
			`{"msg":"EOF","severity":"unknown"}]}`
		if got := jsonString(err, AddWrappedErrors); got != want {
			t.Errorf("got  %s\nwant %s", got, want)
		}
	})

	t.Run("wrapped template", func(t *testing.T) {
		err := errUserLoad.Wrap(errNotFound)
		want := `{"msg":"failed to load user","severity":"unknown","code":"USR-500","wrapped":[` +
			`{"msg":"record not found","severity":"unknown","code":"DB-404"}]}`
		if got := jsonString(err, AddWrappedErrors); got != want {
			t.Errorf("got  %s\nwant %s", got, want)
		}
	})
}

func TestToJSONFields(t *testing.T) {
	err := Template("database unavailable").Code("DB-503").New().
		Set("dsn", "postgres://admin:secret@db").Set("requestId", "r-1")

	tests := []struct {
		name string
		opts []Option
		want string
	}{
		{"no options", nil,
			`{"msg":"database unavailable","severity":"unknown","code":"DB-503"}`},
		{"client", []Option{WithAttributes(ClientOutputFormat)},
			`{"msg":"database unavailable","severity":"unknown","code":"DB-503"}`},
		{"AddFields", []Option{WithAttributes(AddFields)},
			`{"msg":"database unavailable","severity":"unknown","code":"DB-503","fields":{"dsn":"postgres://admin:secret@db","requestId":"r-1"}}`},
		{"root level without AddFields", []Option{WithAttributes(ClientOutputFormat), WithRootLevelFields([]string{"requestId"})},
			`{"msg":"database unavailable","severity":"unknown","code":"DB-503","requestId":"r-1"}`},
		{"root level with AddFields", []Option{WithAttributes(AddFields), WithRootLevelFields([]string{"requestId", "missing"})},
			`{"msg":"database unavailable","severity":"unknown","code":"DB-503","fields":{"dsn":"postgres://admin:secret@db"},"requestId":"r-1"}`},
		{"all fields to root", []Option{WithAttributes(AddFields), WithRootLevelFields([]string{"requestId", "dsn"})},
			`{"msg":"database unavailable","severity":"unknown","code":"DB-503","requestId":"r-1","dsn":"postgres://admin:secret@db"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(ToJSON(err, tt.opts...)); got != tt.want {
				t.Errorf("got  %s\nwant %s", got, tt.want)
			}
			if len(err.fields) != 2 {
				t.Errorf("error fields were modified: %v", err.fields)
			}
		})
	}

	t.Run("Serialize returns a copy of fields", func(t *testing.T) {
		s := Serialize(err, WithAttributes(AddFields))
		s.Fields["dsn"] = "changed"
		if err.fields["dsn"] == "changed" {
			t.Error("error fields were modified through Serialize result")
		}
	})
}

func TestToJSONProtected(t *testing.T) {
	errSQL := Template("pq: relation users does not exist").Code("DB-001").Protected(true)
	errPublic := Template("cannot load profile").Code("PRF-500")

	tests := []struct {
		name  string
		err   error
		rules ErrorSerializationRule
		want  string
	}{
		{"protected skipped without AddProtected",
			errPublic.Wrap(errSQL.Wrap(io.EOF)), AddWrappedErrors,
			`{"msg":"cannot load profile","severity":"unknown","code":"PRF-500","wrapped":[{"msg":"EOF","severity":"unknown"}]}`},
		{"protected written with AddProtected",
			errPublic.Wrap(errSQL.Wrap(io.EOF)), AddWrappedErrors | AddProtected,
			`{"msg":"cannot load profile","severity":"unknown","code":"PRF-500","wrapped":[` +
				`{"msg":"pq: relation users does not exist","severity":"unknown","code":"DB-001"},{"msg":"EOF","severity":"unknown"}]}`},
		{"protected template skipped",
			errPublic.Wrap(errSQL), AddWrappedErrors,
			`{"msg":"cannot load profile","severity":"unknown","code":"PRF-500"}`},
		{"std wrapper with protected inside skipped",
			errPublic.Wrap(fmt.Errorf("query: %w", errSQL.New())), AddWrappedErrors,
			`{"msg":"cannot load profile","severity":"unknown","code":"PRF-500"}`},
		{"std wrapper with protected template inside skipped",
			errPublic.Wrap(fmt.Errorf("query: %w", errSQL)), AddWrappedErrors,
			`{"msg":"cannot load profile","severity":"unknown","code":"PRF-500"}`},
		{"std wrapper with public error inside written",
			errPublic.Wrap(fmt.Errorf("query: %w", Template("timeout").New().Wrap(io.EOF))), AddWrappedErrors,
			`{"msg":"cannot load profile","severity":"unknown","code":"PRF-500","wrapped":[{"msg":"query: timeout: EOF","severity":"unknown"}]}`},
		{"top-level protected written as is",
			errSQL.New(), ClientOutputFormat,
			`{"msg":"pq: relation users does not exist","severity":"unknown","code":"DB-001"}`},
		{"server format writes everything",
			errPublic.Wrap(errSQL.New()), ServerOutputFormat &^ AddStack,
			`{"msg":"cannot load profile","severity":"unknown","code":"PRF-500","wrapped":[{"msg":"pq: relation users does not exist","severity":"unknown","code":"DB-001"}]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jsonString(tt.err, tt.rules); got != tt.want {
				t.Errorf("got  %s\nwant %s", got, tt.want)
			}
		})
	}
}

type failingMarshaler struct{}

func (failingMarshaler) MarshalJSON() ([]byte, error) {
	return nil, se.New(`bad "value"`)
}

func TestToJSONMarshalError(t *testing.T) {
	err := Template("test").New().Set("bad", failingMarshaler{})
	got := ToJSON(err, WithAttributes(AddFields))
	if !json.Valid(got) {
		t.Fatalf("expected valid JSON, got %s", got)
	}
	want := `{"msg":"test","severity":"unknown"}`
	if string(got) != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestToJSONStopStackOn(t *testing.T) {
	err := &Error{metadata: metadata{message: "m"}, stack: []StackFrame{
		{Function: "app.handler"},
		{Function: "app.router"},
		{Function: "github.com/valyala/fasthttp.(*Server).serveConn"},
		{Function: "runtime.main"},
	}}

	tests := []struct {
		stopOn string
		want   []string
	}{
		{"", []string{"app.handler", "app.router", "github.com/valyala/fasthttp.(*Server).serveConn", "runtime.main"}},
		{"fasthttp", []string{"app.handler", "app.router"}},
		{"app.handler", nil},
		{"not-found", []string{"app.handler", "app.router", "github.com/valyala/fasthttp.(*Server).serveConn", "runtime.main"}},
	}

	for _, tt := range tests {
		t.Run(tt.stopOn, func(t *testing.T) {
			s := Serialize(err, WithAttributes(AddStack), WithStopStackOn(tt.stopOn))
			var got []string
			for _, f := range s.Stack {
				got = append(got, f.Function)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
	if len(err.stack) != 4 {
		t.Error("error stack was modified")
	}
}
