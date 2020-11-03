package errors

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

// CapturedError предназначен для обертывания любой ошибки в момент ее
// возникновения для регистрации стека и точки возникновения ошибки.
// для раскрутки иерархии ошибок, необходимо углубляться по Err пока не достигнем
// ошибки отличной от CapturedError.
type CapturedError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
	Code       string `json:"code"`
	frames     []runtime.Frame
	//logmessage []string
	fields   map[string]interface{}
	severity SeverityLevel
	werr     error
	isSecret bool
}

var _ Error = (*CapturedError)(nil)
var _ IsNotFounder = (*CapturedError)(nil)

// New созвращает объект ошибку с текстом msg, одновременно регистрируя точку ее возникновения и стек.
// Используется при отсутствии исходной ошибки, которую необходимо обернуть.
func New(msg string) *CapturedError {
	return newx(msg)
}

func newx(msg string) *CapturedError {
	return &CapturedError{Message: msg, frames: callerFrames(1)}
}

// Capture оборачивает ошибку, одновременно регистрируя точку ее возникновения и стек.
//
// Если ошибка уже является CapturedError, то стек не захватывается, предполагаем что
// это произошло в предыдущем вызове Capture(), значения Severity и StatusCode унаследуются и могут быть
// переопределены позже. Таким образом, возможно развернуть список ошибок сверху-вниз.
// Если err == nil, то ошибка не создается.
func Capture(err error) *CapturedError {
	return capture(err, 0)
}

func capture(err error, callerOffset int) *CapturedError {
	if err == nil {
		return nil
	}

	if ce, ok := err.(*CapturedError); ok {
		res := CapturedError{werr: err}
		res.StatusCode = ce.StatusCode
		res.severity = ce.severity
		return &res
	}

	return &CapturedError{werr: err, frames: callerFrames(callerOffset)}
}

// Error возвращает текст последней ошибки. В скобках указывает количество дочерних ошибок.
func (ce CapturedError) Error() string {
	res := ce.Message
	if len(res) > 0 {
		if d := ce.deep(0); d > 0 {
			res += fmt.Sprintf("[+%d]", d)
		}
		return res
	}
	return ce.werr.Error()
}

func (ce *CapturedError) message() string {
	res := ce.Message
	if len(res) > 0 {
		return res
	}

	if e, ok := ce.werr.(*CapturedError); ok {
		return e.message()
	}

	return ce.Error()
}

func (ce *CapturedError) Children() []Error {
	var res []Error
	unwrap(ce, &res)
	return res
}

func unwrap(ce *CapturedError, dst *[]Error) {
	if e, ok := ce.werr.(*CapturedError); ok {
		*dst = append(*dst, e)
		unwrap(e, dst)
	}
}

func (ce *CapturedError) children() []*CapturedError {
	var res []*CapturedError
	unwrapCE(ce, &res)
	return res
}

func unwrapCE(ce *CapturedError, dst *[]*CapturedError) {
	if e, ok := ce.werr.(*CapturedError); ok {
		*dst = append(*dst, e)
		unwrapCE(e, dst)
	}
}

// CaptureDatabaseError захватывает ошибки связанные с СУБД, инициализирует ключ/значение
// в зависимости от ошибок.
func CaptureDatabaseError(err error) *CapturedError {
	ce := capture(err, 1)
	if pgerr, ok := err.(*pq.Error); ok {
		ce.Set("sqlcode", pgerr.Code)
		if len(pgerr.Constraint) > 0 {
			ce.Set("constraint", pgerr.Constraint)
		}
		if len(pgerr.DataTypeName) > 0 {
			ce.Set("data_type_name", pgerr.DataTypeName)
		}
	}
	return ce
}

func callerFrames(offset int) []runtime.Frame {
	var res []runtime.Frame
	pc := make([]uintptr, 15)
	n := runtime.Callers(3+offset, pc)
	frames := runtime.CallersFrames(pc[:n])

	for {
		frame, more := frames.Next()
		res = append(res, frame)
		if more == false {
			break
		}
	}
	return res
}

func (ce *CapturedError) deep(lvl int) int {
	if e, ok := ce.werr.(*CapturedError); ok {
		return e.deep(lvl + 1)
	}
	return lvl
}

// Msg устанавливает текст сообщения.
func (ce *CapturedError) Msg(s string) Error {
	if ce == nil {
		return nil
	}
	ce.Message = s
	return ce
}

// SetSeverity устанавливает уровень важности сообщения. В соответствии с уровнем важности
// Medium и Critical печатается стэк вызовов функций до места возникновения ошибки.
// Critical - формируется уведомление администратору.
// Если текущий уровень важности, выше level то игнорируется.
func (ce *CapturedError) SetSeverity(level SeverityLevel) Error {
	if ce == nil {
		return nil
	}
	if ce.severity < level {
		ce.severity = level
	}
	return ce
}

// Get возвращает значение по названию ключа.
func (ce *CapturedError) Get(key string) (interface{}, bool) {
	if ce == nil {
		return nil, false
	}

	if ce.fields == nil {
		return nil, false
	}

	res, ok := ce.fields[key]
	return res, ok
}

// SetStatusCode устанавливает HTTP код ошибки, который рекомендовано
// вернуть пользователю.
func (ce *CapturedError) SetStatusCode(code int) Error {
	if ce == nil {
		return nil
	}
	ce.StatusCode = code
	return ce
}

// SetCode устанавливает код ошибки.
func (ce *CapturedError) SetCode(code string) Error {
	if ce == nil {
		return nil
	}
	ce.Code = code
	return ce
}

// Protect исключает ошибку из выдачи в http response.
func (ce *CapturedError) Protect() Error {
	if ce == nil {
		return nil
	}
	ce.isSecret = true
	return ce
}

//
func (ce *CapturedError) IsNotFound() bool {
	if ce == nil {
		return false
	}

	return (ce.StatusCode == 404) && (ce.Message == "not found")
}

func (ce *CapturedError) Log(zl *zerolog.Logger) Error {
	if ce == nil {
		return nil
	}

	e := zl.Error()

	e.Str("severity", ce.severity.String())
	if ce.Code != "" {
		e.Str("errcode", ce.Code)
	}

	if len(ce.fields) > 0 {
		e.Fields(ce.fields)
	}

	if ce.StatusCode != 0 {
		e.Int("status_code", ce.StatusCode)
	}

	arr := ce.children()

	msg := ce.message()
	if len(arr) == 0 {
		e.Str("stack", formatFrames(ce.frames))
		e.Msg(msg)
		return ce
	}

	if len(arr[0].frames) > 0 {
		e.Str("stack", formatFrames(arr[0].frames))
	}

	var arritems []ClientError
	for i := range arr {
		arritems = append(arritems, ClientError{
			Fields:     arr[i].fields,
			Message:    arr[i].Message,
			StatusCode: arr[i].StatusCode,
			Code:       arr[i].Code,
			Severity:   arr[i].severity.String(),
		})
	}

	if len(arr) > 0 {
		e.Interface("errstack", arritems)
	}

	e.Msg(msg)

	return ce
}

func formatFrames(frames []runtime.Frame) string {
	s := ""
	for i := range frames {
		s += frames[i].Function + "() in " + fmt.Sprintf("%s:%d; ", frames[i].File, frames[i].Line)
	}
	return s
}

func (ce *CapturedError) MarshalZerologObject(e *zerolog.Event) {

	e.Str("severity", ce.severity.String())
	e.Str("errcode", ce.Code)
	// errmsg has original error's text.
	e.Str("errmsg", ce.Error())

	if len(ce.fields) > 0 {
		e.Fields(ce.fields)
	}

	if ce.severity == Critical || ce.severity == Medium {
		s := ""
		for i := range ce.frames {
			s += ce.frames[i].Function + "() in " + fmt.Sprintf("%s:%d; ", ce.frames[i].File, ce.frames[i].Line)
		}
		e.Str("stack", s)
	}
	/*
		x := len(ce.logmessage)
		switch {
		case x == 0:
			e.Msg("error captured")
		case x == 1:
			e.Msg(ce.logmessage[0])
		default:
			e.Strs("msgs", ce.logmessage[0:x-1])
			e.Msg(ce.logmessage[x-1])
		}*/
}

// Set сохраняет в ошибке параметры, которые затем будет возможно вывести в журнал.
func (ce *CapturedError) Set(key string, val interface{}) Error {
	if ce == nil {
		return nil
	}

	if ce.fields == nil {
		ce.fields = map[string]interface{}{key: val}
	} else {
		ce.fields[key] = val
	}

	return ce
}

// SetPairs сохраняет в ошибке параметры, которые затем будет возможно вывести в журнал.
func (ce *CapturedError) SetPairs(kvpairs ...interface{}) Error {
	if ce == nil {
		return nil
	}

	if ce.fields == nil {
		ce.fields = map[string]interface{}{}
	}

	for i := 0; i < len(kvpairs)/2; i += 2 {
		ce.fields[fmt.Sprintf("%s", kvpairs[i])] = kvpairs[i+1]
	}

	return ce
}

// SetMulti сохраняет в ошибке параметры, которые затем будет возможно вывести в журнал.
func (ce *CapturedError) SetMulti(key string, vals ...interface{}) Error {
	if ce == nil {
		return nil
	}

	if ce.fields == nil {
		ce.fields = map[string]interface{}{}
	}

	ce.fields[key] = append([]interface{}{}, vals)

	return ce
}

func formatParams(params ...interface{}) string {
	s := "["
	for i := range params {
		s += fmt.Sprintf("%v,", params[i])
	}
	s += "]"
	return s
}

type ClientError struct {
	Message    string                 `json:"message"`
	Code       string                 `json:"code,omitempty"`
	Severity   string                 `json:"severity,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
}

type ClientResponse struct {
	ClientError
	Previous    []ClientError   `json:"previous,omitempty"`
	StackFrames []runtime.Frame `json:"stack_frames,omitempty`
}

func WriterResponse(ctx *fasthttp.RequestCtx, err error, addStackFrames bool) {

	ce, ok := err.(*CapturedError)
	if !ok {
		ctx.SetStatusCode(422)
		ctx.Response.BodyWriter().Write([]byte(fmt.Sprintf(`{"message": "%s"}`, err.Error())))
		return
	}

	ctx.SetStatusCode(ce.StatusCode)
	var cr ClientResponse
	cr.Message = ce.message()
	cr.Code = ce.Code
	cr.Severity = ce.severity.String()
	cr.Fields = ce.fields
	if addStackFrames {
		cr.StackFrames = ce.frames
	}

	ch := ce.children()
	if len(ch) > 0 {
		cr.StackFrames = ch[0].frames
	}

	for i := range ch {
		if ch[i].isSecret {
			continue
		}
		cr.Previous = append(cr.Previous, ClientError{
			Message:  ch[i].message(),
			Code:     ch[i].Code,
			Severity: ch[i].severity.String(),
			Fields:   ch[i].fields,
		})
	}

	enc := json.NewEncoder(ctx.Response.BodyWriter())
	if err := enc.Encode(&cr); err != nil {
		fmt.Println(err.Error())
	}

}
