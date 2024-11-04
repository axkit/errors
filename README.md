# errors 

[![Build Status](https://github.com/axkit/errors/actions/workflows/go.yml/badge.svg)](https://github.com/axkit/errors/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/axkit/errors)](https://goreportcard.com/report/github.com/axkit/errors)
[![GoDoc](https://pkg.go.dev/badge/github.com/axkit/errors)](https://pkg.go.dev/github.com/axkit/errors)
[![Coverage Status](https://coveralls.io/repos/github/axkit/errors/badge.svg?branch=master)](https://coveralls.io/github/axkit/errors?branch=master)


The errors package provides an enterprise approach of error handling.

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