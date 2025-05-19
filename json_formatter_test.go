package errors

import (
	"encoding/json"
	se "errors"
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
	err := &Error{
		metadata: metadata{
			message:    "test error",
			severity:   SeverityLevel(1),
			statusCode: 500,
			code:       "E500",
		},
	}

	options := []Option{
		WithAttributes(AddStack | AddFields),
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
	if jsonResponse.Severity != smedium {
		t.Errorf("expected 'medium', got '%s'", jsonResponse.Severity)
	}
	if jsonResponse.Code != "E500" {
		t.Errorf("expected 'E500', got '%s'", jsonResponse.Code)
	}
	if jsonResponse.StatusCode != 500 {
		t.Errorf("expected 500, got %d", jsonResponse.StatusCode)
	}
}

func TestDefaultErrorMarshaler(t *testing.T) {
	// message := "test error"
	// severity := Critical
	// statusCode := 500
	// code := "E500"

	// jsonBytes := DefaultErrorMarshaler(message, severity, statusCode, code)
	// var jsonResponse map[string]interface{}
	// unmarshalErr := json.Unmarshal(jsonBytes, &jsonResponse)
	// if unmarshalErr != nil {
	// 	t.Errorf("unexpected error: %v", unmarshalErr)
	// }
	// if jsonResponse["message"] != message {
	// 	t.Errorf("expected '%s', got '%s'", message, jsonResponse["message"])
	// }
	// if jsonResponse["severity"] != scritical {
	// 	t.Errorf("expected 'critical', got '%s'", jsonResponse["severity"])
	// }
	// if jsonResponse["statusCode"] != float64(statusCode) {
	// 	t.Errorf("expected %d, got %f", statusCode, jsonResponse["statusCode"])
	// }
	// if jsonResponse["code"] != code {
	// 	t.Errorf("expected '%s', got '%s'", code, jsonResponse["code"])
	// }
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

func TestToJSONWithContext(t *testing.T) {
	embeddedError := Template("embedded error").Code("E1234").StatusCode(400)
	err := &Error{
		metadata: metadata{
			message:    "test error",
			severity:   Medium,
			statusCode: 500,
			code:       "E500",
		},
		err:   embeddedError,
		stack: DefaultCallerFrames(0),
	}

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
  "severity": "tiny"
}`

	if response := string(ToJSON(err, WithAttributes(IndentJSON))); response != expectedResponse {
		t.Errorf("expected '%s', got '%s'", expectedResponse, response)
	}
}
