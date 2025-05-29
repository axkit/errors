package errors

import (
	"encoding/json"
	se "errors"
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

		if len(jsonResponse.Stack) != 1 {
			t.Errorf("expected 1 stack frame, got %d", len(jsonResponse.Stack))
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

		if len(jsonResponse.Stack) != 1 {
			t.Errorf("expected 1 stack frame, got %d", len(jsonResponse.Stack))
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
