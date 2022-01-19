package errors

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type FormattingFlag uint8

const (

	// AddStack - add stack in the JSON.
	AddStack FormattingFlag = 1 << iota

	// AddProtected - add protected errors in the JSON.
	AddProtected

	// AddProtected - add key/value pairs.
	AddFields

	// AddWrappedErrors - add to the output previous errors.
	AddWrappedErrors
)

// ToJSON formats error as JSON with flags.
func ToJSON(err error, flags FormattingFlag) []byte {
	return err2json(err, flags)
}

// ToClientJSON returns error formatted as JSON addressed to HTTP response.
// {
// 	"msg" : "validation failed",
// 	"code" : "ORA-0600",
// 	"severity" : "critical",
// 	"statusCode" : 400
// }
func ToClientJSON(err error) []byte {
	return err2json(err, 0)
}

// ToServerJSON returns error formatted as JSON addressed to server logs.
func ToServerJSON(err error) []byte {
	return err2json(err, AddProtected|AddStack|AddFields|AddWrappedErrors)
}

func err2json(err error, flags FormattingFlag) []byte {

	if err == nil {
		return nil
	}

	ce, ok := err.(*CatchedError)
	if !ok {
		return []byte(`{"msg": "` + err.Error() + `"}`)
	}

	buf := bytes.NewBuffer([]byte(`{"msg":"`))
	buf.WriteString(ce.lasterr.Message)
	buf.Write([]byte(`","severity":"`))
	buf.WriteString(ce.lasterr.Severity.String())
	buf.WriteByte('"')

	if ce.lasterr.Code != "" {
		buf.Write([]byte(`,"code":"`))
		buf.WriteString(ce.lasterr.Code)
		buf.WriteByte('"')
	}
	if ce.lasterr.StatusCode != 0 {
		buf.WriteString(`,"statusCode":`)
		buf.WriteString(strconv.Itoa(ce.lasterr.StatusCode))
	}

	if len(ce.fields) > 0 {
		for _, key := range RootLevelFields {
			v, ok := ce.fields[key]
			if ok {
				buf.WriteByte(',')
				buf.WriteString(kvString(key, v))
				delete(ce.fields, key)
			}
		}
	}

	if len(ce.fields) > 0 && flags&AddFields == AddFields {
		buf.Write([]byte(`,"ctx":{`))
		sep := byte(0)
		for k, v := range ce.fields {
			if sep != 0 {
				buf.WriteByte(sep)
			}
			buf.WriteString(kvString(k, v))
			sep = ','
		}
		buf.WriteByte('}')
	}

	if flags&AddWrappedErrors == AddWrappedErrors && len(ce.werrs) > 0 {
		buf.Write([]byte(`,"errs":[`))
		sep := byte(0)
		for i := range ce.werrs {
			if ce.werrs[i].Protected {
				continue
			}
			if sep != 0 {
				buf.WriteByte(sep)
			}
			buf.WriteString(we2json(&ce.werrs[i]))
			sep = ','
		}
		buf.WriteByte(']')
	}

	if flags&AddStack == AddStack && len(ce.frames) > 0 {
		buf.Write([]byte(`,"stack":"`))
		sep := []byte{}
		for i := range ce.frames {
			if strings.Contains(ce.frames[i].Function, CaptureStackStopWord) {
				break
			}
			if strings.Contains(ce.frames[i].Function, "github.com/axkit/errors") {
				continue
			}
			buf.Write(sep)
			buf.WriteString(ce.frames[i].Function)
			buf.Write([]byte(`() in `))
			buf.WriteString(ce.frames[i].File)
			buf.WriteByte(':')
			buf.WriteString(strconv.Itoa(ce.frames[i].Line))
			sep = []byte(`; `)
		}
		buf.WriteByte('"')
	}

	buf.WriteByte('}')

	return buf.Bytes()
}

// we2json converts WrapperError
func we2json(we *WrappedError) string {

	res := `{"severity":"` + we.Severity.String() + `"`
	if we.Code != "" {
		res += `,"code": "` + we.Code + `"`
	}

	res += `,"msg":"` + strings.Replace(we.Message, "\"", "'", -1) + `"`
	if we.StatusCode != 0 {
		res += `,"statusCode":` + strconv.Itoa(we.StatusCode)
	}
	res += "}"

	return res
}

// kvString returns key and value as json key value pair.
func kvString(key string, val interface{}) string {
	var res string
	if s, ok := val.(interface{ String() string }); ok {
		return `"` + key + `":"` + s.String() + `"`
	}

	if s, ok := val.([]string); ok {
		res = `"` + key + `":[`
		sep := ""
		for i := range s {
			res += sep + `"` + s[i] + `"`
			sep = ","
		}
		res += "]"
		return res
	}

	if s, ok := val.(string); ok {
		return `"` + key + `":"` + s + `"`
	}

	if b, ok := val.([]byte); ok {
		return `"` + key + `":` + string(b)
	}

	return fmt.Sprintf(`"%s":%v`, key, val)
}
