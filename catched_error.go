package errors

import (
	"errors"
	"fmt"
)

// WrappedError is a structure defining wrapped error.
type WrappedError struct {

	// Message holds final error's Message.
	Message string `json:"message"`

	// Severity) holds severity level.
	Severity SeverityLevel `json:"severity"`

	// StatusCode holds HTTP status code, what is recommended to assign to HTTP response
	// if value is specified (above zero).
	StatusCode int `json:"-"`

	// Code holds application error code.
	Code string `json:"code,omitempty"`

	// Protected
	Protected bool `json:"-"`

	// err is wrapped original error.
	err               error
	isMessageReplaced bool
}

// Error implements standard Error interface.
func (we WrappedError) Error() string {
	return we.Message
}

// Err returns original error.
func (we WrappedError) Err() error {
	if we.err != nil {
		return we.err
	}
	return errors.New(we.Message)
}

// CatchedError holds call stack
type CatchedError struct {
	Frames  []Frame                `json:"frames,omitempty"`
	Fields  map[string]interface{} `json:"-"`
	lasterr WrappedError
	werrs   []WrappedError
}

// New returns CatchedError with stack at the point of calling with severity level Tiny.
// The function is used if there is no original error what can be wrapped.
func New(msg string) *CatchedError {
	return newx(msg)
}

// NewMedium returns CatchedError with stack at the point of calling and severity level Medium.
func NewMedium(msg string) *CatchedError {
	return newx(msg).Severity(Medium)
}

// NewCritical returns CatchedError with stack at the point of calling and severity level Critical.
func NewCritical(msg string) *CatchedError {
	return newx(msg).Severity(Critical)
}

func newx(msg string) *CatchedError {
	return &CatchedError{lasterr: WrappedError{Message: msg, isMessageReplaced: true}, Frames: DefaultCallerFramesFunc(1)}
}

// Catch wraps an error with capturing stack at the point of calling.
// A severity level is Tiny.
//
// If err is already CapturedError, the stack does not capture again. It's assumed
// it was done before. The attributes Severity and StatusCode inherits but can be changed later.
//
// if err == nil returns nil.
func Catch(err error) *CatchedError {
	if err == nil {
		return nil
	}
	return catch(err, 0)
}

// CatchCustom wraps an error with custom stack capturer.
func CatchCustom(err error, stackcapture func() []Frame) *CatchedError {
	if err == nil {
		return nil
	}

	return &CatchedError{lasterr: WrappedError{Message: err.Error(), err: err, isMessageReplaced: true}, Frames: stackcapture()}
}

func catch(err error, callerOffset int) *CatchedError {

	if ce, ok := err.(*CatchedError); ok {
		ce.werrs = append(ce.werrs, ce.lasterr)
		ce.lasterr.Protected = false
		ce.lasterr.isMessageReplaced = false
		return ce
	}

	return &CatchedError{
		lasterr: WrappedError{
			Message:           err.Error(),
			err:               err,
			isMessageReplaced: true,
		},
		Frames: DefaultCallerFramesFunc(callerOffset),
	}
}

// Msg replaces error message.
func (ce *CatchedError) Msg(s string) *CatchedError {
	if ce == nil {
		return nil
	}
	ce.lasterr.Message = s
	ce.lasterr.isMessageReplaced = true
	return ce
}

// Last returns last wrapper.
func (ce *CatchedError) Last() *WrappedError {
	return &ce.lasterr
}

// Len returns amount of errors wrapped + 1.
func (ce *CatchedError) Len() int {
	return 1 + len(ce.werrs)
}

// Error implements interface golang/errors Error.  Returns string, taking into
// account value of ErrorMethodMode.
//
// If ErrorMethodMode == Multi, results build using LIFO principle.
func (ce CatchedError) Error() string {
	if ErrorMethodMode == Single {
		if ce.lasterr.Message != "" {
			return ce.lasterr.Message
		}

		for i := len(ce.werrs) - 1; i >= 0; i-- {
			if res := ce.werrs[i].Message; res != "" {
				return res
			}
		}
		return "?"
	}

	// Multi
	res := ce.lasterr.Message
	for i := len(ce.werrs) - 1; i >= 0; i-- {
		if msg := ce.werrs[i].Message; msg != "" {
			res += ": " + msg
		}
	}

	return res
}

// Strs returns error messages of wrapped errors except last message and empty messages.
func (ce *CatchedError) Strs(exceptProtected bool) []string {
	if len(ce.werrs) == 0 {
		return nil
	}

	res := make([]string, 0, len(ce.werrs))

	for i := len(ce.werrs) - 1; i >= 0; i-- {
		if ce.werrs[i].Protected && exceptProtected {
			continue
		}
		if msg := ce.werrs[i].Message; msg != "" && ce.werrs[i].isMessageReplaced {
			res = append(res, msg)
		}
	}
	return res
}

// CaptureDatabaseError захватывает ошибки связанные с СУБД, инициализирует ключ/значение
// в зависимости от ошибок.
// func CaptureDatabaseError(err error) *CapturedError {
// 	ce := capture(err, 1)
// 	if pgerr, ok := err.(*pq.Error); ok {
// 		ce.Set("sqlcode", pgerr.Code)
// 		if len(pgerr.Constraint) > 0 {
// 			ce.Set("constraint", pgerr.Constraint)
// 		}
// 		if len(pgerr.DataTypeName) > 0 {
// 			ce.Set("data_type_name", pgerr.DataTypeName)
// 		}
// 	}
// 	return ce
// }

// Severity set error's severity level. It's ignored if level
// is lower than current level.
func (ce *CatchedError) Severity(level SeverityLevel) *CatchedError {
	if ce == nil {
		return nil
	}
	if ce.lasterr.Severity < level {
		ce.lasterr.Severity = level
	}
	return ce
}

// Medium sets severity level to Medium.
func (ce *CatchedError) Medium() *CatchedError {
	return ce.Severity(Medium)
}

// Critical sets severity level to Critical.
func (ce *CatchedError) Critical() *CatchedError {
	return ce.Severity(Critical)
}

// Get returns value by key.
func (ce *CatchedError) Get(key string) (interface{}, bool) {
	if ce == nil {
		return nil, false
	}

	if ce.Fields == nil {
		return nil, false
	}

	res, ok := ce.Fields[key]
	return res, ok
}

// GetDefault returns value by key. Returns def if not found.
func (ce *CatchedError) GetDefault(key string, def interface{}) interface{} {
	if ce == nil {
		return def
	}

	if ce.Fields == nil {
		return def
	}

	res, ok := ce.Fields[key]
	if !ok {
		return def
	}
	return res
}

// StatusCode sets HTTP response code, recommended to be assigned.
func (ce *CatchedError) StatusCode(code int) *CatchedError {
	if ce == nil {
		return nil
	}
	ce.lasterr.StatusCode = code
	return ce
}

// Code sets business code of.
func (ce *CatchedError) Code(code string) *CatchedError {
	if ce == nil {
		return nil
	}
	ce.lasterr.Code = code
	return ce
}

// Protect marks error as protected. An error with protection will not be
// visible to the user.
func (ce *CatchedError) Protect() *CatchedError {
	if ce == nil {
		return nil
	}
	ce.lasterr.Protected = true
	return ce
}

// Set accociates a single key with value.
func (ce *CatchedError) Set(key string, val interface{}) *CatchedError {
	if ce == nil {
		return nil
	}

	if ce.Fields == nil {
		ce.Fields = map[string]interface{}{}
	}

	ce.Fields[key] = val
	return ce
}

// LastNonPairedValue holds value to be assigned by SetPairs if amount of parameters is odd.
var LastNonPairedValue interface{} = "missed value"

// SetPairs accociates multiple key/value pairs. SetPairs("id", 10, "name", "John")
// if amount of parameters is odd, SetPairs("id", 10, "name") uses LastNonPairedValue
// as the last value.
func (ce *CatchedError) SetPairs(kvpairs ...interface{}) *CatchedError {
	if ce == nil {
		return nil
	}

	if len(kvpairs) == 0 {
		return ce
	}

	even := true
	if len(kvpairs)%2 != 0 {
		even = false
	}

	if ce.Fields == nil {
		ce.Fields = map[string]interface{}{}
	}

	for i := 0; i < len(kvpairs)/2; i += 2 {
		key := fmt.Sprintf("%s", kvpairs[i])
		var val interface{}
		if !even && i == len(kvpairs)-1 {
			val = LastNonPairedValue
		} else {
			val = kvpairs[i+1]
		}
		ce.Fields[key] = val
	}
	return ce
}

// SetVals accociates with the key multiple interfaces.
func (ce *CatchedError) SetVals(key string, vals ...interface{}) *CatchedError {
	if ce == nil {
		return nil
	}

	if ce.Fields == nil {
		ce.Fields = map[string]interface{}{}
	}

	ce.Fields[key] = vals
	return ce
}

// SetStrs accociates with the key multiple strings.
func (ce *CatchedError) SetStrs(key string, strs ...string) *CatchedError {
	if ce == nil {
		return nil
	}

	if ce.Fields == nil {
		ce.Fields = map[string]interface{}{}
	}

	ce.Fields[key] = strs
	return ce
}

// IsNotFound returns true if it's not found error wrapped.
func (ce *CatchedError) IsNotFound() bool {
	if ce == nil {
		return false
	}

	return (ce.lasterr.StatusCode == 404) && (ce.lasterr.Message == "not found")
}
