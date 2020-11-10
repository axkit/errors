package errors_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/axkit/errors"
)

func f() {
	e := errors.New("invalid rules")
	fmt.Println("e.Error()=", e.Error(), e.Len())

	es := errors.Catch(e).Msg("end of file reached").Set("file", "config.json").Severity(errors.Critical)
	fmt.Println("es.Error()=", es.Error(), e.Len())

	ess := errors.Catch(es).Msg("processing failed").Set("process", "billing")
	fmt.Println("ess.Error()=", ess.Error(), e.Len())

	errors.ErrorMethodMode = errors.Multi
	fmt.Println("ess.Error()=", ess.Error(), ess.GetDefault("file", ""), e.Len())

	//fmt.Println(ess.Err())

	errors.ErrorMethodMode = errors.Single

	js := errors.Catch(es).Msg("file reading error")
	fmt.Println("js.Error()=", js.Error())

	//fmt.Println(js.Log())

	// arr := js.Children()
	// for i := range arr {
	// 	fmt.Printf("%d = %#v\n", i, arr[i])
	// }
	// fmt.Println("")
	// fmt.Printf("e=%#v\n", e)

}
func TestSeverityLevel_MarshalJSON(t *testing.T) {

	tc := []struct {
		level    errors.SeverityLevel
		expected []byte
	}{
		{level: errors.Tiny, expected: []byte(`"tiny"`)},
		{level: errors.Medium, expected: []byte(`"medium"`)},
		{level: errors.Critical, expected: []byte(`"critical"`)},
		{level: 15, expected: []byte(`"unknown"`)},
	}

	for i := range tc {
		buf, err := json.Marshal(tc[i].level)
		if err != nil {
			t.Errorf("#1 failed case %d. Details: %s", i, err.Error())
		}
		if !bytes.Equal(tc[i].expected, buf) {
			t.Errorf("#2 failed case %d. Expected  %s, got  %s", i, string(tc[i].expected), string(buf))
		}
	}
}

func BenchmarkSeverityLevel_MarshalJSON(b *testing.B) {
	tc := []struct {
		level    errors.SeverityLevel
		expected []byte
	}{
		{level: errors.Tiny, expected: []byte(`"tiny"`)},
		{level: errors.Medium, expected: []byte(`"medium"`)},
		{level: errors.Critical, expected: []byte(`"critical"`)},
		{level: 15, expected: []byte(`"unknown"`)},
	}

	for i := 0; i < b.N; i++ {
		k := i % 4
		buf, err := json.Marshal(tc[k].level)
		if err != nil {
			b.Error(err)
		}
		if !bytes.Equal(tc[k].expected, buf) {
			b.Errorf("#2 failed case %d. Expected  %s, got  %s", i, string(tc[i].expected), string(buf))
		}
		_ = len(buf)
	}
}
