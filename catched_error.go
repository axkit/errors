package errors

import (
	"errors"
	"fmt"
)

// WrappedError is a structure defining wrapped error. It's public to be able
// customize logging and failed HTTP responses.
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

	// err holds initial error.
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

// CatchedError holds original error, all intermediate error wraps and the call stack.
type CatchedError struct {
	frames       []Frame
	fields       map[string]interface{}
	lasterr      WrappedError
	werrs        []WrappedError
	isRecatched  bool
	isStackAdded bool
}

// CatchedError implements golang standard Error interface. Returns string, taking into
// account setting ErrorMethodMode.
//
// If ErrorMethodMode == Multi, results build using LIFO principle.
func (ce CatchedError) Error() string {
	res := ce.lasterr.Message
	if ErrorMethodMode == Single {
		return res
	}

	// Multi
	for i := len(ce.werrs) - 1; i >= 0; i-- {
		if msg := ce.werrs[i].Message; msg != "" {
			res += ": " + msg
		}
	}

	return res
}

// Frames returns callstack frames.
func (ce *CatchedError) Frames() []Frame {
	return ce.frames
}

// Fields returns all key/value pairs associated with error.
func (ce *CatchedError) Fields() map[string]interface{} {
	return ce.fields
}

// Last returns the last wrap of error .
func (ce *CatchedError) Last() WrappedError {
	return ce.lasterr
}

// Len returns amount of errors wrapped + 1 or zero if nil.
func (ce *CatchedError) Len() int {
	if ce == nil {
		return 0
	}
	return 1 + len(ce.werrs)
}

// WrappedErrors returns all errors holding by CatchedError.
func (ce *CatchedError) WrappedErrors() []WrappedError {
	res := make([]WrappedError, 1+len(ce.werrs))
	res[0] = ce.lasterr
	if len(ce.werrs) > 0 {
		copy(res[1:], ce.werrs)
	}
	return res
}

// Alarmer wraps a single method Alarm that receives CatchedError and implement real-time
// notification logic.
//
// The type CapturedError has method Alarm that recieves Alarmer as a parameter.
type Alarmer interface {
	Alarm(*CatchedError)
}

// New returns *CatchedError with stack at the point of calling and  severity level Tiny.
// The function is used if there is no original error what can be wrapped.
func New(msg string) *CatchedError {
	return newx(msg, true)
}

// NewTiny is synonym to func New.
func NewTiny(msg string) *CatchedError {
	return newx(msg, true)
}

// NewMedium returns *CatchedError with stack at the point of calling and severity level Medium.
func NewMedium(msg string) *CatchedError {
	return newx(msg, true).Severity(Medium)
}

// NewCritical returns CatchedError with stack at the point of calling and severity level Critical.
func NewCritical(msg string) *CatchedError {
	return newx(msg, true).Severity(Critical)
}

func newx(msg string, stack bool) *CatchedError {
	res := CatchedError{lasterr: WrappedError{Message: msg, isMessageReplaced: false}}
	if stack {
		res.frames = CallerFramesFunc(1)
	}
	return &res
}

// Catch wraps an error with capturing stack at the point of calling.
// Assigns severity level Tiny.
//
// If err is already CapturedError, the stack does not capture again. It's assumed
// it was done before. The attributes Severity and StatusCode inherits but can be changed later.
//
// Returns nil if err is nil.
func Catch(err error) *CatchedError {
	if err == nil {
		return nil
	}
	return catch(err, 0)
}

func Wrap(err error, ce *CatchedError) *CatchedError {
	if err == nil {
		return nil
	}

	e := catch(err, 0)

	if len(ce.fields) > 0 {
		if len(e.fields) == 0 {
			e.fields = ce.fields
		} else {
			for k, v := range ce.fields {
				e.fields[k] = v
			}
		}
	}

	e.werrs = append(e.werrs, e.lasterr)
	e.lasterr = ce.lasterr

	return e
}

// Raise returns explicitly defined CatchedError. Function captures stack at the point of calling.
func Raise(ce *CatchedError) *CatchedError {
	ce.frames = CallerFramesFunc(1)
	ce.isStackAdded = true
	return ce
}

// CatchCustom wraps an error with custom stack capturer.
func CatchCustom(err error, stackcapture func() []Frame) *CatchedError {
	if err == nil {
		return nil
	}

	return &CatchedError{
		lasterr:      WrappedError{Message: err.Error(), err: err, isMessageReplaced: false},
		frames:       stackcapture(),
		isStackAdded: true}
}

func catch(err error, callerOffset int) *CatchedError {

	if ce, ok := err.(*CatchedError); ok {
		// message still stay the same!	It's expected message will be replaced later by calling Msg().
		(*ce).isRecatched = true
		return ce
	}

	return &CatchedError{
		lasterr: WrappedError{
			Message: err.Error(),
			err:     err,
		},
		frames:       CallerFramesFunc(callerOffset),
		isStackAdded: true,
	}
}

// Capture captures stack frames. Recommended to use when raised predefined errors.
//
// var ErrInvalidCustomerID = errors.New("invalid customer id")
//
// if c, ok := customers[id]; ok {
// 	  return ErrInvalidCustomerID.Capture()
// }
func (ce *CatchedError) Capture() *CatchedError {
	ce.frames = CallerFramesFunc(0)
	ce.isStackAdded = true
	return ce
}

// Msg sets or replaces latest error's text message.
// If message different previous error pushed to error stack.
func (ce *CatchedError) Msg(s string) error {
	if ce == nil {
		return nil
	}

	if ce.lasterr.Message == s {
		return ce
	}

	ce.werrs = append(ce.werrs, ce.lasterr)
	ce.lasterr.Message = s
	return ce
}

// WrappedMessages returns error messages of wrapped errors except last message.
func (ce *CatchedError) WrappedMessages(exceptProtected bool) []string {
	return ce.messages(false, exceptProtected)
}

// AllMessages returns all error text messages including top (last) message.
// The last message is in the beginning of slice.
func (ce *CatchedError) AllMessages(exceptProtected bool) []string {
	return ce.messages(true, exceptProtected)
}

func (ce *CatchedError) messages(includeTop, exceptProtected bool) []string {
	var res []string

	if includeTop {
		if !(exceptProtected && ce.lasterr.Protected) && ce.lasterr.Message != "" {
			res = append(res, ce.lasterr.Message)
		}
	}

	for i := len(ce.werrs) - 1; i >= 0; i-- {
		if ce.werrs[i].Protected && exceptProtected {
			continue
		}
		res = append(res, ce.werrs[i].Message)
	}

	return res
}

// Severity overwrites error's severity level.
func (ce *CatchedError) Severity(level SeverityLevel) *CatchedError {
	if ce == nil {
		return nil
	}

	ce.lasterr.Severity = level

	return ce
}

func (ce *CatchedError) GetSeverity() SeverityLevel {
	return ce.lasterr.Severity
}

func (ce *CatchedError) GetCode() string {
	return ce.lasterr.Code
}

// Medium sets severity level to Medium. It's ignored if current level Critical.
func (ce *CatchedError) Medium() *CatchedError {
	return ce.Severity(Medium)
}

// Critical sets severity level to Critical.
func (ce *CatchedError) Critical() *CatchedError {
	return ce.Severity(Critical)
}

// StatusCode sets HTTP response code, recommended to be assigned.
// StatusCode 404 is used by
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

// Set associates a single key with value.
func (ce *CatchedError) Set(key string, val interface{}) *CatchedError {
	if ce == nil {
		return nil
	}

	if ce.fields == nil {
		ce.fields = map[string]interface{}{}
	}

	ce.fields[key] = val
	return ce
}

// LastNonPairedValue holds value to be assigned by SetPairs if amount of parameters is odd.
var LastNonPairedValue interface{} = "missed value"

// SetPairs associates multiple key/value pairs. SetPairs("id", 10, "name", "John")
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

	if ce.fields == nil {
		ce.fields = map[string]interface{}{}
	}

	for i := 0; i < len(kvpairs); i += 2 {
		key := fmt.Sprintf("%s", kvpairs[i])
		var val interface{}
		if !even && i == len(kvpairs)-1 {
			val = LastNonPairedValue
		} else {
			val = kvpairs[i+1]
		}
		ce.fields[key] = val
	}
	return ce
}

// SetVals accociates with the key multiple interfaces.
func (ce *CatchedError) SetVals(key string, vals ...interface{}) *CatchedError {
	if ce == nil {
		return nil
	}

	if ce.fields == nil {
		ce.fields = map[string]interface{}{}
	}

	ce.fields[key] = vals
	return ce
}

// SetStrs accociates with the key multiple strings.
func (ce *CatchedError) SetStrs(key string, strs ...string) *CatchedError {
	if ce == nil {
		return nil
	}

	if ce.fields == nil {
		ce.fields = map[string]interface{}{}
	}

	ce.fields[key] = strs
	return ce
}

// Get returns value by key.
func (ce *CatchedError) Get(key string) (interface{}, bool) {
	if ce == nil {
		return nil, false
	}

	if ce.fields == nil {
		return nil, false
	}

	res, ok := ce.fields[key]
	return res, ok
}

// GetDefault returns value by key. Returns def if not found.
func (ce *CatchedError) GetDefault(key string, def interface{}) interface{} {
	if ce == nil {
		return def
	}

	if ce.fields == nil {
		return def
	}

	res, ok := ce.fields[key]
	if !ok {
		return def
	}
	return res
}

// IsNotFound returns true if StatusCode is 404.
func (ce *CatchedError) IsNotFound() bool {
	return IsNotFound(ce)
}

// IsNotFound returns true if err is *CatchedError and StatusCode is 404.
// If err wraps another errors, last one is taken for decision.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	if ce, ok := err.(*CatchedError); ok {
		return ce.lasterr.StatusCode == 404
	}

	return false
}

// Alarm send error to Alarmer.
// Intended usage is real-time SRE notification if critical error.
func (ce *CatchedError) Alarm(a Alarmer) {
	a.Alarm(ce)
}
