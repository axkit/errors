package errors

/*
import (
	"github.com/rs/zerolog"
	"runtime"
)

var (
	ErrNotFound         = HandlerError{RecommendedStatusCode: 440, BasicError: BasicError{Code: "ERR-0440", Message: "объект не найден!"}}
	ErrBadRequest       = HandlerError{RecommendedStatusCode: 400, BasicError: BasicError{Code: "ERR-0400", Message: "некорректные параметры вызова!"}}
	ErrInternalError    = HandlerError{RecommendedStatusCode: 500, BasicError: BasicError{Code: "ERR-0500", Message: "внутреняя ошибка!", Severity: Critical}}
	ErrInvalidReference = HandlerError{RecommendedStatusCode: 500, BasicError: BasicError{Code: "ERR-0500", Message: "нарушение ссылочной целостности!", Severity: Critical}}
	ErrSQLError         = BasicError{Code: "DB-0500", Message: "ошибка при выполнении SQL запроса", Severity: Critical}
)

type Error interface {
	error
	Basic() BasicError
	MarshalZerologObject(e *zerolog.Event)
}

// BasicError описывает структуру базового класса ошибки.
type BasicError struct {
	Code           string          `json:"code"`
	Message        string          `json:"message"`
	Severity       SeverityLevel   `json:"severity"`
	Kind           Kind            `json:"kind"`
	Errors         []error         `json:"errors"`
	Frames         []runtime.Frame `json:"frames"`
	Op             string          `json:"op"`
	Recommendation string          `json:"recommendation"`
}

// WrappedError .
type WrappedError interface {
	LogDetails() map[string]string
}

// Wrap оборачивает .
func Wrap(be Error, err error) error {

	b := be.Basic()
	b.Errors = append(b.Errors, err)
	process(&b)

	return b
}

// Raise
func Raise(err Error) error {

	b := err.Basic()
	process(&b)

	return err
}

func process(be *BasicError) {

	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])

	for {
		frame, more := frames.Next()
		be.Frames = append(be.Frames, frame)
		if more == false {
			break
		}
	}
}

// SeverityLevel описывает возможные уровни серьезности ошибок.
type SeverityLevel int

const (

	// Tiny классифицирует ожидаемые, возможные ошибки, не требующие привлечения внимания администратора.
	Tiny SeverityLevel = iota

	// Medium классифицируют ошибку средней категории, требующее регистрации в специальном журнале и/или
	// отображения на дашборде.
	Medium

	// Critical классифицирует критическую ошибку, информация о возникновении подобной ошибки должна немедленно
	// передаваться администратору всеми возможными доступными способами.
	Critical
)

func (esl SeverityLevel) String() string {
	switch esl {
	case Tiny:
		return "tiny"
	case Medium:
		return "medium"
	case Critical:
		return "critical"
	}
	return ""
}

// Kind описывает классы ошибок по типам.
type Kind int8

const (
	Other   Kind = iota // Unclassified error. This value is not printed in the error message.
	Invalid             // Invalid operation for this type of item.
	BadRequest
	Permission // Permission denied.
	IO         // External I/O error such as network failure.
	Exist      // Item already exists.
	NotExist   // Item does not exist.
	//	IsDir           // Item is a directory.
	//	NotDir          // Item is not a directory.
	//	NotEmpty        // Directory not empty.
	Internal        // Internal error or inconsistency.
	BrokenReference // Link target does not exist.
)

func (k Kind) String() string {
	switch k {
	case Other:
		return "other error"
	case Invalid:
		return "invalid operation"
	case BadRequest:
		return "bad request"
	case Permission:
		return "permission denied"
	case IO:
		return "I/O error"
	case Exist:
		return "item already exists"
	case NotExist:
		return "item does not exist"
	case BrokenReference:
		return "link target does not exist"
		/*	case IsDir:
				return "item is a directory"
			case NotDir:
				return "item is not a directory"
			case NotEmpty:
				return "directory not empty"
			case Private:
				return "information withheld"*/
/*	case Internal:
	return "internal error"
	/*case Transient:
	return "transient error"*/
/*	}
	return "unknown error kind"
}

func (be BasicError) Error() string {
	return be.Code + ": " + be.Message
}

func (be BasicError) Basic() BasicError {
	return be
}
func (be BasicError) MarshalZerologObject(e *zerolog.Event) {
	e.Str("code", be.Code)
	if len(be.Errors) > 0 {
		e.Str("msg", be.Errors[0].Error())
	}
	// e.Msg(be.Message)
}

// HandlerError описывает типовую ошибку которую возвращает Handler.
type HandlerError struct {
	BasicError
	RecommendedStatusCode int `json:"-"`
}

type DatabaseError struct {
	BasicError
	TableName       string `json:"table_name"`
	OriginalCode    string `json:"original_code"`
	OriginalMessage string `json:"original_message"`
}

func (e HandlerError) Error() string {
	return e.Code + ": " + e.Message
}


// NotificationLevel represents how fast system administrator should be
// notified about happend error.
type NotificationLevel int

const (
	None NotificationLevel = iota
	Regular
	Immediately
)


type ErrorNotifier interface {
	Notify(Error)
}

type Error struct {
	// Code unique
	Code           string
	DefaulHTTPCode int
	Kind           int

	MessageUID string

	// Op holds operation name. Usually includes pkg/function name.
	Op string
}

// me holds all predefined errors created usign New() method.
var me = make(map[string]Error)

func init() {
	Init()
}

// Raise returns error with alerting if required.
func Raise(err Error) Error {
	return err
}
*/
