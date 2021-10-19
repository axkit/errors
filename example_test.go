package errors_test

import (
	"fmt"
	"strconv"

	"github.com/axkit/errors"
)

func ExampleNew() {

	var err error
	var ErrValidationFailed = errors.New("value validation failed").StatusCode(400)

	if _, err = strconv.Atoi("5g"); err != nil {
		err = errors.Wrap(err, ErrValidationFailed.Set("value", "5g"))
	}

	//err := errors.Raise(ErrAuthServiceIsNotAvailable).StatusCode(500).Critical()
	fmt.Println(err.Error())
}
