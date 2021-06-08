package errors_test

import (
	"fmt"

	"github.com/axkit/errors"
)

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

func ExampleNew() {

	var ErrAuthServiceIsNotAvailable = errors.New("auth service not available")

	//
	// if res, ok := client.SendRequest(login, password); ok {
	//	return
	// }

	err := errors.Raise(ErrAuthServiceIsNotAvailable).StatusCode(500).Critical()
	fmt.Println(err.Error())
	// Output:
	// auth service not available
}
