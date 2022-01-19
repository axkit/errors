package errors

import (
	"strings"
	"testing"
)

func TestToServerJSON(t *testing.T) {

	expected := `"msg":"unknown customer","severity":"medium","statusCode":404,"ctx":{"hello":{"customerId":1,"typ":"organization"}}`

	err := NotFound("unknown customer").Set("hello", []byte(`{"customerId":1,"typ":"organization"}`))

	if s := string(ToServerJSON(err)); !strings.Contains(s, expected) {
		t.Error(s)
	}
}
