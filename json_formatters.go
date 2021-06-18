package errors

import (
	"fmt"
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
)

func ToJSON(err error, flags FormattingFlag) []byte {
	return err2json(err, flags)
}

// JSON returns JSON representation of error.
func ToClientJSON(err error) []byte {
	return err2json(err, 0)
}

func ToServerJSON(err error) []byte {
	return err2json(err, AddProtected|AddStack|AddFields)
}

// JSON4Client returns JSON representation of error as following object
// {
// 	"msg" : "text",
// 	"code" : "ORA-0600",
// 	"severity" : "critical",
// 	"statusCode" :
// }
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
			res += fmt.Sprintf(`,"%s":%v`, k, v)
		}
	}

	if ce.lasterr.StatusCode != 0 {
		res += fmt.Sprintf(`,"statusCode":%d`, ce.lasterr.StatusCode)
	}

	if len(ce.werrs) > 0 {
		res += `,"errs":[`
		sep := ""
		for i := range ce.werrs {
			res += fmt.Sprintf(`%s{"msg":"%s"}`, sep, ce.werrs[i].Message)
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
