package errors_test

import (
	"fmt"
	"io"
	"strconv"
	"testing"
	"time"

	"github.com/axkit/errors"
)

func TestCatch(t *testing.T) {

	err := errors.Catch(io.ErrUnexpectedEOF)

	if err.Last().Message != io.ErrUnexpectedEOF.Error() {
		t.Errorf("#1 failed. Expected %s, got %s", io.ErrUnexpectedEOF.Error(), err.Last().Message)
	}

	msgs := err.AllMessages(false)
	if msgs[0] != io.ErrUnexpectedEOF.Error() {
		t.Errorf("#2 failed. Expected %s, got %s", io.ErrUnexpectedEOF.Error(), msgs[0])
	}

	errx := errors.Catch(err)
	msgs = errx.AllMessages(false)
	if msgs[0] != io.ErrUnexpectedEOF.Error() {
		t.Errorf("#3 failed. Expected %s, got %s", io.ErrUnexpectedEOF.Error(), msgs[0])
	}

	if errx.Len() != 1 {
		t.Error("#4 failed. Wrapping was not expected")
	}
}

func TestCatchedError_Severity(t *testing.T) {

	err := errors.Catch(io.ErrUnexpectedEOF)
	if err.GetSeverity() != errors.Tiny {
		t.Error("#1 failed")
	}

	if s := err.Medium().GetSeverity(); s != errors.Medium {
		t.Errorf("#2 failed. Expected %s, got %s", errors.Medium, s)
	}

	if s := err.Severity(errors.Critical).GetSeverity(); s != errors.Critical {
		t.Errorf("#3 failed. Expected %s, got %s", errors.Critical, s)
	}

	err = errors.NewCritical("critical error")
	if s := err.GetSeverity(); s != errors.Critical {
		t.Errorf("#4 failed. Expected %s, got %s", errors.Critical, s)
	}
}

func TestCatchedError_Code(t *testing.T) {

	err := errors.Catch(io.ErrUnexpectedEOF)
	if err.GetCode() != "" {
		t.Error("#1 failed")
	}

	if s := err.Code("ORA-1234").GetCode(); s != "ORA-1234" {
		t.Errorf("#2 failed. Expected %s, got %s", "ORA-1234", s)
	}

	err = errors.NewCritical("critical error").Code("ORA-0600")
	if s := err.GetCode(); s != "ORA-0600" {
		t.Errorf("#3 failed. Expected %s, got %s", "ORA-0600", s)
	}
}

func TestCatchedError_SetPairs(t *testing.T) {

	err := errors.Catch(io.ErrUnexpectedEOF).SetPairs("filename", "data.csv", "size", 4096)

	m := err.Fields()
	if len(m) != 2 {
		t.Errorf("#1 failed. Expected 2 fields, got %d", len(m))
	}

	found := 0
	for k, v := range m {
		switch k {
		case "filename":
			if v != "data.csv" {
				t.Errorf("#2 failed. Expected %s, got %s", "data.csv", v)
			}
			found++
		case "size":
			if v != 4096 {
				t.Errorf("#3 failed. Expected %d, got %d", 4096, v)
			}
			found++
		}
	}

	if found != 2 {
		t.Error("#4 failed. Not all key-pairs retrived")
	}
}

func TestCatchedError_SetPairsMissed(t *testing.T) {

	err := errors.Catch(io.ErrUnexpectedEOF).SetPairs("filename", "data.csv", "size")

	m := err.Fields()
	if len(m) != 2 {
		t.Errorf("#1 failed. Expected 2 fields, got %d", len(m))
	}

	found := 0
	for k, v := range m {
		switch k {
		case "filename":
			if v != "data.csv" {
				t.Errorf("#2 failed. Expected %s, got %s", "data.csv", v)
			}
			found++
		case "size":
			if v != errors.LastNonPairedValue {
				t.Errorf("#3 failed. Expected %s, got %v", errors.LastNonPairedValue, v)
			}
			found++
		}
	}

	if found != 2 {
		t.Error("#4 failed. Not all key-pairs retrived")
	}
}
func TestCatchedError_SetPairsEmpty(t *testing.T) {

	err := errors.Catch(io.ErrUnexpectedEOF).SetPairs()

	m := err.Fields()
	if m != nil {
		t.Errorf("Failed. Expected empty fields (nil), got %d elements", len(m))
	}
}

var ts = []struct {
	key string
	val interface{}
}{{"id", int(42)}, {"name", "John"}, {"at", time.Now()}}

func TestCatchedError_Fields(t *testing.T) {
	err := errors.Catch(io.ErrUnexpectedEOF)

	for i := range ts {
		err.Set(ts[i].key, ts[i].val)
	}

	m := err.Fields()
	if m == nil {
		t.Error("#1 failed. Expected non empty fields, got nil")
		t.FailNow()
	}

	if len(m) != len(ts) {
		t.Errorf("#2 failed. Different fields length, expected %d got %d", len(ts), len(m))
		t.FailNow()
	}

	for i := range ts {
		v, ok := m[ts[i].key]
		if !ok {
			t.Errorf("#3 failed, case %d. Not found \"%s\"", i, ts[i].key)
		}

		if v != ts[i].val {
			t.Errorf("#4 failed, case %d. Value not equal. Expected %v, got %v", i, v, ts[i].val)
		}

	}
}

func TestCatchedError_Set(t *testing.T) {

	err := errors.Catch(io.ErrUnexpectedEOF)

	for i := range ts {
		m := err.Set(ts[i].key, ts[i].val).Fields()
		if len(m) != i+1 {
			t.Errorf("#1 failed, case: %d. Key/value not added", i)
		}
	}

	for i := range ts {
		v, ok := err.Get(ts[i].key)
		if !ok {
			t.Errorf("#3 failed, case %d. Not found \"%s\"", i, ts[i].key)
		}

		if v != ts[i].val {
			t.Errorf("#4 failed, case %d. Value not equal. Expected %v, got %v", i, v, ts[i].val)
		}
	}
}
func TestCatchedError_Get(t *testing.T) {

	ts := []struct {
		key string
		val interface{}
	}{{"id", int(42)}, {"name", "John"}, {"at", time.Now()}}

	err := errors.Catch(io.ErrUnexpectedEOF)

	v, ok := err.Get("non-existing-key")
	if ok || v != nil {
		t.Errorf("#1 failed. Expected nil, got %v", v)
		t.FailNow()
	}

	for i := range ts {
		m := err.Set(ts[i].key, ts[i].val).Fields()
		if len(m) != i+1 {
			t.Errorf("#1 failed, case: %d. Key/value not added", i)
		}
	}

	m := err.Fields()
	if m == nil {
		t.Error("#2 failed. Expected non empty fields, got nil")
		t.FailNow()
	}

	for i := range ts {
		v, ok := err.Get(ts[i].key)
		if !ok {
			t.Errorf("#3 failed, case %d. Not found \"%s\"", i, ts[i].key)
		}

		if v != ts[i].val {
			t.Errorf("#4 failed, case %d. Value not equal. Expected %v, got %v", i, v, ts[i].val)
		}
	}
}

func TestIsNotFound(t *testing.T) {

	err := errors.Catch(io.ErrUnexpectedEOF)
	if errors.IsNotFound(err) || errors.IsNotFound(nil) {
		t.Error("#1 failed. Expected false, got true")
	}

	err.StatusCode(404)

	if !errors.IsNotFound(err) {
		t.Error("#2 failed. Expected true, got false")
	}
}

func TestIs(t *testing.T) {

	err1 := errors.New("not found")
	err2 := err1

	if !errors.Is(err1, err2) {
		t.Error("case #1")
	}

	err3 := errors.NotFound("not found")
	err4 := err3
	if !errors.Is(err3, err4) {
		t.Error("case #2")
	}

	err5 := errors.NotFound("not found")
	err6 := err3.Raise()
	if !errors.Is(err5, err6) {
		t.Error("case #3")
	}

}

func TestCatchedError_Error(t *testing.T) {

	errors.ErrorMethodMode = errors.Single
	err := errors.Catch(io.ErrUnexpectedEOF)

	if err.Error() != io.ErrUnexpectedEOF.Error() {
		t.Error("#1 failed. Error messages are not equal")
	}

	errors.ErrorMethodMode = errors.Multi

	txt := "file reading failed"
	expected := txt + ": " + io.ErrUnexpectedEOF.Error()
	errx := err.Msg(txt)
	if msg := errx.Error(); expected != msg {
		t.Errorf("#2 failed. Error messages are not equal, expected %s, got %s", expected, msg)
	}
}

func TestToClientJSON(t *testing.T) {

	err := errors.Catch(io.ErrUnexpectedEOF).SetPairs("id", 42, "name", "John").StatusCode(500).Code("ORA-0600").Msg("file reading failed")

	fmt.Printf("%s\n", string(errors.ToClientJSON(err)))
}

func TestToServerJSON(t *testing.T) {

	err := errors.Catch(io.ErrUnexpectedEOF).SetPairs("id", 42, "name", "John").StatusCode(500).Code("ORA-0600").Msg("file reading failed")

	fmt.Printf("%s\n", string(errors.ToServerJSON(err)))
}

/*
func f() {
	e := errors.New("invalid rules")
	fmt.Println("e.Error()=", e.Error(), e.Len())

	es := errors.Catch(e).Msg("end of file reached").Set("file", "config.json").Critical()
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

func TestCatchedError_Strs(t *testing.T) {

	msg := []string{"original/protected", "intermediate", "user faced"}

	err1 := errors.New(msg[0]).Protect()
	err2 := errors.Catch(err1).Msg(msg[1]).StatusCode(401)
	err3 := errors.Catch(err2).Msg(msg[2])

	{
		s := err3.WrappedMessages(false)
		exp := []string{msg[1], msg[0]}
		if !deepEqual(exp, s) {
			t.Errorf("#1 failed. expected %v, got %v", exp, s)
		}
	}
	{
		s := err3.WrappedMessages(true)
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

func jsonToMap(buf []byte) (map[string]interface{}, error) {

	res := map[string]interface{}{}
	err := json.Unmarshal(buf, &res)
	if err != nil {
		return nil, errors.Catch(err).Msg("byte representation of CatchedError unmarshal failed")
	}

	return res, nil
}

func TestCatchedError_Msg(t *testing.T) {

	msg := "end of file"
	err := errors.Catch(io.ErrUnexpectedEOF).Msg(msg)

	if s := err.Last().Message; s != msg {
		t.Errorf("test failed, expected %s, got %s", msg, s)
	}

}
func TestCatchedError_JSONPart(t *testing.T) {

	msg := "end of file"
	err := errors.Catch(io.ErrUnexpectedEOF).Msg(msg).StatusCode(500)

	j, errx := jsonToMap(errors.JSONPart(err, errors.NoStack))
	if errx != nil {
		t.Error(errx)
	}

	if j["msg"] != msg {
		t.Errorf("test failed, expected %s, got %v", msg, j["msg"])
	}

	if sc, ok := j["statusCode"]; !ok || int64(sc.(float64)) != 500 {
		t.Errorf("test failed, expected %d, got %v", 500, sc)
	}

	if _, ok := j["stack"]; ok {
		t.Error("test failed, found not expected attr \"stack\"")
	}
}

func TestCatchedError_JSON(t *testing.T) {

	msg := "end of file"
	err := errors.Catch(io.ErrUnexpectedEOF).Msg(msg).StatusCode(500).Set("key", "val")

	j, errx := jsonToMap(errors.JSON(err))
	if errx != nil {
		t.Error(errx)
	}

	if j["msg"] != msg {
		t.Errorf("test failed, expected %s, got %v", msg, j["msg"])
	}

	if sc, ok := j["statusCode"]; !ok || int64(sc.(float64)) != 500 {
		t.Errorf("test failed, expected %d, got %v", 500, sc)
	}

	if _, ok := j["stack"]; !ok {
		t.Error("test failed, not found expected attr \"stack\"")
	}

	if v, ok := j["key"]; !ok || v.(string) != "val" {
		t.Errorf("test failed, expected %s, got %v", "val", v)
	}

}

func TestCatchedError_WrappedMessages(t *testing.T) {

	msg := "end of file"
	err1 := errors.Catch(io.ErrUnexpectedEOF)
	fmt.Printf("err1 %v\n", strings.Join(err1.AllMessages(false), "->"))
	err1.Msg("hola")
	fmt.Printf("err1 %v\n", strings.Join(err1.AllMessages(false), "->"))

	err11 := errors.Catch(err1)
	fmt.Printf("err11 %v\n", strings.Join(err11.AllMessages(false), "->"))

	err12 := errors.Catch(err11).Msg("12")
	fmt.Printf("err12 %v\n", strings.Join(err12.AllMessages(false), "->"))

	err2 := errors.Catch(io.ErrUnexpectedEOF).Msg(msg)
	fmt.Printf("err2 %v\n", strings.Join(err2.AllMessages(false), "->"))

	err21 := errors.Catch(err2)
	fmt.Printf("err21 %v\n", strings.Join(err21.AllMessages(false), "->"))
	err22 := errors.Catch(err2).Msg("22")
	fmt.Printf("err22 %v\n", strings.Join(err22.AllMessages(false), "->"))
	err22.Msg("xx")
	fmt.Printf("err22 %v\n", strings.Join(err22.AllMessages(false), "->"))

	//	fmt.Printf("%v\n", err.Last())
	//s	fmt.Printf("%v\n", err.WrappedErrors())
	/*
		j, errx := jsonToMap(errors.JSON(err))

		if errx != nil {
			t.Error(errx)
		}

		if j["msg"] != msg {
			t.Errorf("test failed, expected %s, got %v", msg, j["msg"])
		}

		if sc, ok := j["statusCode"]; !ok || int64(sc.(float64)) != 500 {
			t.Errorf("test failed, expected %d, got %v", 500, sc)
		}

		if _, ok := j["stack"]; !ok {
			t.Error("test failed, not found expected attr \"stack\"")
		}

		if v, ok := j["key"]; !ok || v.(string) != "val" {
			t.Errorf("test failed, expected %s, got %v", "val", v)
		}*/

//}

func TestWrap(t *testing.T) {

	errors.ErrorMethodMode = errors.Single
	var err error
	var ErrValidationFailed = errors.New("value validation failed").StatusCode(400)

	if _, err = strconv.Atoi("5g"); err != nil {
		err = errors.Wrap(err, ErrValidationFailed.Set("value", "5g"))
	}

	if err.Error() != "value validation failed" {
		t.Error("expected another message, got:", err.Error())
	}

	t.Log(string(errors.ToJSON(err, errors.AddWrappedErrors)))
	t.Log(string(errors.ToJSON(err, 0)))

	//err := errors.Raise(ErrAuthServiceIsNotAvailable).StatusCode(500).Critical()
}

func TestCatchedError_FieldasStringSlice(t *testing.T) {
	err := errors.Catch(io.ErrUnexpectedEOF)

	for i := range ts {
		err.Set(ts[i].key, ts[i].val)
	}
	err.Set("perms", []string{"Login", "CreateUser", "UpdateUser"})

	m := err.Fields()
	if m == nil {
		t.Error("#1 failed. Expected non empty fields, got nil")
		t.FailNow()
	}

	t.Logf("%s", string(errors.ToServerJSON(err)))
}

func TestCatchedError_JSONFormatting(t *testing.T) {

	err1 := errors.ValidationFailed("validation failed").Code("001")
	err2 := errors.InternalError().Code("002")

	ae := errors.Wrap(io.ErrUnexpectedEOF, err1)
	be := errors.Wrap(ae, err2)
	ae.Set("reason", "hello world")
	ae.Set("lala", time.Now())

	// err := errors.Catch(io.ErrUnexpectedEOF).Code("ORA-0600").Set("userId", 98)
	// err.Set("perms", []string{"Login", "CreateUser", "UpdateUser"})

	// xe := errors.Catch(err).Set("name", "Robert").Code("AAA-0001")

	// t.Logf("%s", string(errors.ToJSON(xe, errors.AddStack|errors.AddProtected|errors.AddFields|errors.AddWrappedErrors)))

	// be := errors.Wrap(io.ErrUnexpectedEOF, xe).Code("XE-001")
	t.Logf("%s", string(errors.ToJSON(be, errors.AddStack|errors.AddProtected|errors.AddFields|errors.AddWrappedErrors)))

}
