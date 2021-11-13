package errors

import (
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

	res := "{"
	res += fmt.Sprintf(`"msg":"%s"`, ce.lasterr.Message)
	res += fmt.Sprintf(`,"severity":"%s"`, ce.lasterr.Severity.String())

	if ce.lasterr.Code != "" {
		res += fmt.Sprintf(`,"code":"%s"`, ce.Last().Code)
	}

	if len(ce.fields) > 0 && flags&AddFields == AddFields {
		for k, v := range ce.fields {
			if s, ok := v.(interface{ String() string }); ok {
				res += fmt.Sprintf(`,"%s":"%s"`, k, s.String())
				continue
			}
			if s, ok := v.([]string); ok {
				res += `,"` + k + `": [`
				sep := ""
				for i := range s {
					res += fmt.Sprintf(`%s"%s"`, sep, s[i])
					sep = ","
				}
				res += "]"
				continue
			}

			if s, ok := v.(string); ok {
				res += fmt.Sprintf(`,"%s":"%s"`, k, s)
				continue
			}

			res += fmt.Sprintf(`,"%s":%v`, k, v)
		}
	}

	if ce.lasterr.StatusCode != 0 {
		res += fmt.Sprintf(`,"statusCode":%d`, ce.lasterr.StatusCode)
	}

	if len(ce.werrs) > 0 && flags&AddWrappedErrors == AddWrappedErrors {
		res += `,"errs":[`
		sep := ""
		for i := range ce.werrs {
			res += sep + we2json(&ce.werrs[i])
			sep = ","
		}
		res += "]"
	}

	if flags&AddStack == AddStack {
		s := ""
		for i := range ce.frames {
			if strings.Contains(ce.frames[i].Function, CaptureStackStopWord) {
				break
			}
			if strings.Contains(ce.frames[i].Function, "github.com/axkit/errors") {
				continue
			}
			s += ce.frames[i].Function + "() in " + fmt.Sprintf("%s:%d;", ce.frames[i].File, ce.frames[i].Line)
		}
		res += fmt.Sprintf(`,"stack":"%s"`, s)
	}

	res += "}"
	return []byte(res)
}

// we2json converts WrapperError
func we2json(we *WrappedError) string {

	if we.Protected {
		return ""
	}

	res := "{"
	sep := ""
	if we.Code != "" {
		res += `"code": "` + we.Code + `"`
		sep = ","
	}

	res += sep + `"severity":"` + we.Severity.String() + `"`
	sep = ","
	res += sep + `"msg":"` + we.Message
	if we.StatusCode != 0 {
		res += sep + `"statusCode":` + strconv.Itoa(we.StatusCode)
	}
	res += "}"

	return res
}
