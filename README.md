# errors 

[![Build Status](https://github.com/axkit/errors/actions/workflows/go.yml/badge.svg)](https://github.com/axkit/errors/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/axkit/errors)](https://goreportcard.com/report/github.com/axkit/errors)
[![GoDoc](https://pkg.go.dev/badge/github.com/axkit/errors)](https://pkg.go.dev/github.com/axkit/errors)
[![Coverage Status](https://coveralls.io/repos/github/axkit/errors/badge.svg?branch=master)](https://coveralls.io/github/axkit/errors?branch=master)

The errors package provides an enterprise-grade error handling approach.

## Overview
This Go package provides a robust framework for error handling with advanced features tailored for enterprise-grade applications. It introduces structured error templates with attributes like severity, status codes, and more, ensuring both clarity and utility when diagnosing issues in production environments.

### Key Features
- **Structured Error Attributes**: Add metadata like severity level, application-specific codes, and HTTP status codes.
- **Error Wrapping**: Seamlessly wrap errors to maintain stack trace and additional context.
- **Predefined Errors**: Declare reusable error templates for consistency.
- **JSON Formatting**: Serialize errors into JSON differently for API responses and logging.
- **Severity Levels**: Classify errors as `Tiny`, `Medium`, or `Critical` for proper handling.
- **Stack Frames**: Capture detailed call stacks for debugging.
- **Alarmer Interface**: Automatically notify administrators of critical errors.


## Installation
```bash
go get github.com/axkit/errors
```


## Error Structure
The `Error` type is the core of this package. It encapsulates metadata, stack traces, and wrapped errors.

### Attributes
| Attribute   | Description                                     |
|-------------|-------------------------------------------------|
| `message`   | Error message text                             |
| `severity`  | Severity of the error (Tiny, Medium, Critical) |
| `statusCode`| HTTP status code                               |
| `code`      | Application-specific error code                |
| `protected` | Indicates that error's attributes shall not leak to the client.            |
| `fields`    | Custom key-value pairs for additional context  |
| `stack`     | Stack frames showing the call trace            |


## Error Template
Predefined errors offer reusable templates for consistent error creation. Use the `New` function to declare them:

```go

import "github.com/axkit/errors"

var (
    ErrInvalidInput = errors.New("invalid input provided").
							Code("VAL-0901").
							StatusCode(400).
							Severity(errors.Tiny)
    
	ErrDatabaseDown = errors.New("database is unreachable").
							Code("DBA-0253").
							StatusCode(500).
							Severity(errors.Critical)

	// Predefined error gets `Tiny` severity  by default.
	ErrInvalidFilter = errors.New("invalid filtering rule").
							Code("VAL-0900").
							StatusCode(400)

)

if err := db.Ping(ctx); err != nil {
	return ErrDatabaseDown.Wrap(err)
}


```




## Examples

### Basic Error Creation
The same way as standard error. Can be extended by metadata.

```go
package main

import (
    "fmt"
    "github.com/axkit/errors"
)

func main() {

	var err1 = errors.New("something went wrong")
    var err2 = errors.New("something went wrong").Code("E1001").Severity(errors.Medium)
    
	fmt.Println(err1.Error()) // something went wrong
	fmt.Println(err2.Error()) // something went wrong
	fmt.Println(errors.JSON(err)) // 

}
```

### Wrapping Errors
```go
wrappedErr := errors.New("Outer error").Wrap(errors.New("Inner error"))
fmt.Println(wrappedErr.Error())
```

### Using Predefined Errors
```go
if err := someDatabaseOperation(); err != nil {
    return ErrDatabaseDown.Wrap(err)
}
```

```go

import "github.com/axkit/errors"

var ErrObjectNotFound = errors.New("object not found").Code("CMN-0404")
							
func (r *CustomerRepo)CustomerByID(customerID int) (CustomerDto, error)  
   customer, ok := customers[customerID]
   if !ok {
		return ErrObjectNotFound.Raise().Set("customerID", customerID)
   }

   return customer, nil 
}

// 
// ...
// 
func (s *CustomerService)CustomerByID(customerID int, requestedBy string) (*Customer, error) {

	c, err := s.repo.CustomerByID(customerID)
	if err != nil {
		return nil, errors.Wrap(err, "customer not found").Set("requestedBy", requestedBy).Sererity(errors.Medium)
	}

	return convertDto(c), nil 
}

//
// ... 
//
var ErrCustomerNotFound = errors.New("customer not found").Code("CRM-0404")
func (s *CustomerService)CustomerBySSN(ssn string, requestedBy string) (*Customer, error) {

	c, err := s.repo.CustomerBySSN(ssn)
	if err != nil {
		return nil, ErrorCustomerNotFound.Wrap(err)
	}

	return convertDto(c), nil 
}
 
```
### JSON Formatting
```go
jsonErr := errors.New("Resource not found").Code("E404").StatusCode(404)
fmt.Println(string(errors.ToJSON(jsonErr, errors.WithAttributes(errors.AddStack))))
```

### Alarm Notifications
Set an alarmer to notify on critical errors: 
```go
type CustomAlarmer struct{}

func (c *CustomAlarmer) Alarm(err error) {
    fmt.Println("Critical error:", err)
}

errors.SetAlarmer(&CustomAlarmer{})
errors.New("message queue connection failure").Severity(errors.Critical).Alarm()
```

## Why Use This Package?

- **Support simplification**: Assigning a business code to each error allows for detailed documentation of possible causes and solutions for resolving them.   
- **Consistency**: Predefined errors enforce uniform error handling across the codebase.
- **Improved Observability**: Structured errors with stack traces and metadata provide clarity in debugging.
- **JSON Serialization**: Convert errors to JSON for external logging or API error responses.
- **Actionable Severity Levels**: Severity classification helps prioritize critical issues.
- **Extensibility**: Use custom alarmers to integrate with monitoring systems.

## Severity Levels
The package classifies errors into three severity levels:
- **Tiny**: Minor issues, typically validation errors.
- **Medium**: Regular errors that log stack traces.
- **Critical**: Major issues requiring immediate attention.

---

## Contributing
Contributions are welcome! Please submit issues or pull requests on [GitHub](https://github.com/axkit/errors).

---

## License
This project is licensed under the MIT License.











In large and complex applications, each error typically has a unique code. Using this code, you can find in the documentation the possible causes of the error and the ways to resolve it. This package allows you to define a list of errors with specific attributes: error code, message text, corresponding HTTP status code, and severity level.

Example:
```go 
var ErrInvalidCustomerID = errors.New("invalid customer id").
    StatusCode(400).
    Code("ERR-0199").
    Severity(errors.Medium)
```
With a unique error code, we can direct the user to the page
`https://mysomeservice.com/content/errors/ERR-0199`
where the causes of the error and methods for resolving it will be described in detail, if necessary.

Any predefined error is instantiated and transformed into an Error object when the Raise method is invoked.
Additionally, key-value pairs can be attached to the error, which can be viewed by the administrator in the logs,
all predefined attributes cab be reassigned.

```go
    // type request struct {
	//		SessionID int 
	//		CustomerID int 
	// 		CustomerFirstName string  
	// }

	if request.CustomerID <= 0 {
		return ErrInvalidCustomerID.Raise().
			Set("sessionId", request.SessionID).
			Severity(errors.Critical)
	}
```