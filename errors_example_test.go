package errors_test

import (
	"fmt"

	"github.com/axkit/errors"
)

func ExampleNew_one() {

	var ErrInvalidInput = errors.New("invalid input").
		Code("CMN-0400").
		StatusCode(400).
		Severity(errors.Tiny)

	fmt.Println(ErrInvalidInput)
	// Output: invalid input
}

func ExampleNew_two() {

	var ErrSimpleError = errors.New("simple error")

	fmt.Println(ErrSimpleError.Error())
	// Output: simple error
}

func ExampleError_Error() {

	type Input struct {
		ID        int    `json:id"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}

	var ErrEmptyAttribute = errors.New("empty attribute value").Code("CMN-0400")
	var ErrInvalidInput = errors.New("invalid input").Code("CMN-0400")

	validateInput := func(inp *Input) error {
		if inp.ID == 0 {
			return ErrEmptyAttribute.Build().Set("emptyFields", []string{"id"})
		}
		return nil
	}

	if err := validateInput(&Input{}); err != nil {
		returnErr := ErrInvalidInput.Wrap(err)
		fmt.Printf(returnErr.Error())
		// Output: invalid input: empty attribute value
	}
}

// ExampleToJSON demonstrates generating JSON output for an error.
func ExampleToJSON() {
	jsonErr := errors.New("User not found").Code("E404").StatusCode(404).Severity(errors.Tiny)
	jsonOutput := errors.ToJSON(jsonErr, errors.WithAttributes(errors.AddFields))
	fmt.Println("JSON Error:", string(jsonOutput))
	// Output: JSON Error: {"msg":"User not found","severity":"tiny","code":"E404","statusCode":404}
}

// ExampleWrap demonstrates wrapping an error.
func ExampleWrap() {
	innerErr := errors.New("Database connection failed")
	outerErr := errors.New("Service initialization failed").Wrap(innerErr)
	fmt.Println("Wrapped Error:", outerErr.Error())
	// Output: Wrapped Error: Service initialization failed: Database connection failed
}

// ExamplePredefinedErrors demonstrates using predefined errors.
func ExampleErrorTemplate() {
	var ErrDatabaseDown = errors.New("Database is unreachable").
		Code("DB-500").
		StatusCode(500).
		Severity(errors.Critical)

	if err := openDatabase("pg:5432"); err != nil {
		fmt.Println("Error:", ErrDatabaseDown.Wrap(err).Error())
		// Output: Error: Database is unreachable: unable to connect to database
	}
}

// ExampleAlarm demonstrates raising an alarm for critical errors.

type CustomAlarmer struct{}

func (c *CustomAlarmer) Alarm(err error) {
	fmt.Println("Critical error:", err)
}

func ExampleAlarmer() {

	errors.SetAlarmer(&CustomAlarmer{})
	var ErrSystemFailure = errors.New("system failure").Severity(errors.Critical)

	ErrSystemFailure.Build().Set("path", "/var/lib").Alarm()

	// Output: Critical error: system failure
}

func openDatabase(connStr string) error {
	var dbErr error
	return errors.Wrap(dbErr, "unable to connect to database").Set("connectionString", connStr)
}
