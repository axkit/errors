package errors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"
)

func f() {
	e := New("invalid rules")
	fmt.Println("e.Error()=", e.Error(), e.Len())

	es := Catch(e).Msg("end of file reached").Set("file", "config.json").Severity(Critical)
	fmt.Println("es.Error()=", es.Error(), e.Len())

	ess := Catch(es).Msg("processing failed").Set("process", "billing")
	fmt.Println("ess.Error()=", ess.Error(), e.Len())

	ErrorMethodMode = Multi
	fmt.Println("ess.Error()=", ess.Error(), ess.GetDefault("file", ""), e.Len())

	//fmt.Println(ess.Err())

	ErrorMethodMode = Single

	js := Catch(es).Msg("file reading error")
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
		level    SeverityLevel
		expected []byte
	}{
		{level: Tiny, expected: []byte(`"tiny"`)},
		{level: Medium, expected: []byte(`"medium"`)},
		{level: Critical, expected: []byte(`"critical"`)},
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
		level    SeverityLevel
		expected []byte
	}{
		{level: Tiny, expected: []byte(`"tiny"`)},
		{level: Medium, expected: []byte(`"medium"`)},
		{level: Critical, expected: []byte(`"critical"`)},
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

func TestCatchedError_Strs(t *testing.T) {

	msg := []string{"original/protected", "intermediate", "user faced"}

	err1 := New(msg[0]).Protect()
	err2 := Catch(err1).Msg(msg[1]).StatusCode(401)
	err3 := Catch(err2).Msg(msg[2])

	{
		s := err3.Strs(false)
		exp := []string{msg[1], msg[0]}
		if !deepEqual(exp, s) {
			t.Errorf("#1 failed. expected %v, got %v", exp, s)
		}
	}
	{
		s := err3.Strs(true)
		exp := []string{msg[1]}
		if !deepEqual(exp, s) {
			t.Errorf("#2 failed. expected %v, got %v", exp, s)
		}
	}

}

func deepEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}
func TestCatchedError_Msg(t *testing.T) {

	msg := "end of file"
	err := Catch(io.ErrUnexpectedEOF).Msg(msg)

	if s := err.Last().Message; s != msg {
		t.Errorf("test failed, expected %s, got %s", msg, s)
	}

}
func TestCatchedError_JSON(t *testing.T) {

	msg := "end of file"
	err := Catch(io.ErrUnexpectedEOF).Msg(msg)

	t.Log(string(JSON(err)))

	t.Log(string(JSON(io.ErrUnexpectedEOF)))
}
