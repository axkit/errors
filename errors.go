package errors

import (
	"database/sql"

	"github.com/rs/zerolog"
)

// SeverityLevel описывает уровень серьезности ошибки.
type SeverityLevel int

const (
	// Tiny классифицирует ожидаемые, возможные ошибки, не требующие привлечения внимания администратора.
	// В журнал не пишется стек вызовов функций. Пример: ошибка валидации введенных полей формы.
	Tiny SeverityLevel = iota

	// Medium классифицируют ошибку средней категории, требующее регистрации в специальном журнале и/или
	// отображения на дашборде. В журнал пишется стек вызовов функций.
	Medium

	// Critical классифицирует критическую ошибку, информация о возникновении подобной ошибки должна немедленно
	// передаваться администратору всеми возможными доступными способами. В журнал пишется стек вызовов функций.
	Critical
)

// String возвращает строковое представление уровня серьезности ошибки. В английском языке
func (sl SeverityLevel) String() string {
	switch sl {
	case Tiny:
		return "tiny"
	case Medium:
		return "medium"
	case Critical:
		return "critical"
	}
	return "unknown"
}

// Error описывает интерфейс враппера ошибок. Прозрачная замена для пакета errors из стандартной библиотеки.
type Error interface {

	// Set привязывает ключ/значение к ошибке.
	Set(key string, val interface{}) Error

	// SetPairs привязывает несколько пар ключ/значение к ошибке. Количество параметров должно быть кратное двум.
	// где нечетный имеет тип строку: название ключа, второй значение любого типа.
	SetPairs(kvpairs ...interface{}) Error

	// SetMulti принимает ключ и список значений. Преимущественно для привязки параметров SQL запроса.
	SetMulti(key string, vals ...interface{}) Error

	Get(key string) (interface{}, bool)
	SetSeverity(SeverityLevel) Error
	SetStatusCode(int) Error
	SetCode(string) Error

	// Msg устанавливает текст ошибки.
	Msg(string) Error

	// Error обеспечивает совместимость с error.
	Error() string

	// Children возвращает дочерние ошибки. Текущая ошибка в массив не попадает!
	Children() []Error

	Protect() Error

	Log(*zerolog.Logger) Error
}

type IsNotFounder interface {
	IsNotFound() bool
}

// IsNotFound возвращает true если ошибка связана с отсутствием данных.
// Если err == nil, возвращает false. Для ошибки sql.ErrNoRows возвращает true.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	if err == sql.ErrNoRows {
		return true
	}

	if nferr, ok := err.(IsNotFounder); ok {
		return nferr.IsNotFound()
	}
	return false
}

// NotFound объект не найден.
var NotFound = func(msg string) Error {
	return newx(msg).SetStatusCode(404).SetCode("RDK-0404").SetSeverity(Medium)
}

var ValidationFailed = func(msg string) Error {
	return newx(msg).SetStatusCode(400).SetCode("RDK-0400").SetSeverity(Tiny)
}

var ConsistencyFailed = func() Error {
	return newx("consistency failed").SetStatusCode(500).SetSeverity(Critical)
}

var InvalidRequestBody = func(s string) Error {
	return newx(s).SetStatusCode(400).SetSeverity(Critical)
}

var Unauthorized = func() Error {
	return newx("unauthorized").SetStatusCode(401).SetSeverity(Medium)
}

var Forbidden = func() Error {
	return newx("forbidden").SetStatusCode(403).SetSeverity(Critical)
}

var InternalError = func() Error {
	return newx("internal error").SetStatusCode(500).SetSeverity(Critical)
}

var UnprocessableEntity = func(s string) Error {
	return newx(s).SetStatusCode(422).SetSeverity(Medium)
}
