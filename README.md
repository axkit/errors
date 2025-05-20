# errors

[![Build Status](https://github.com/axkit/errors/actions/workflows/go.yml/badge.svg)](https://github.com/axkit/errors/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/axkit/errors)](https://goreportcard.com/report/github.com/axkit/errors)
[![GoDoc](https://pkg.go.dev/badge/github.com/axkit/errors)](https://pkg.go.dev/github.com/axkit/errors)
[![Coverage Status](https://coveralls.io/repos/github/axkit/errors/badge.svg?branch=master)](https://coveralls.io/github/axkit/errors?branch=master)

The errors package provides an enterprise-grade error handling approach.

## Motivation

### 1. Unique Error Codes for Client Support

In many corporate applications, it’s not enough for a system to return an error - it must also return a unique, user-facing error code. This code allows end users or support engineers to reference documentation or helpdesk pages that explain what the error means, under what conditions it might occur, and how to resolve it.

These codes serve as stable references and are especially valuable in systems where frontend, backend, and support operations must all stay synchronized on error semantics.

### 2. Centralized Error Definitions

To make error codes reliable and consistent, they must be centrally defined and maintained in advance. Without such coordination, teams risk introducing duplicate codes, inconsistent messages, or undocumented behaviors. This package was built to support structured error registration and reuse across the entire application.

### 3. Logging at the Top Level & Contextual Information in Errors

There is an important idiom: log errors only once, and do it as close to the top of the call stack as possible — for instance, in an HTTP controller. Lower layers (business logic, database, etc.) may wrap and propagate errors upward, but only the outermost layer should produce the log entry.

This pattern, while clean and idiomatic, introduces a challenge: how can we include rich, contextual information at the logging point, if it was only known deep inside the application?

To solve this, we need errors that are context-aware. That is, they should carry structured attributes — like the IP address of a failing server, or the input that triggered the issue — as they move up the call stack. This package provides facilities to attach such structured context to errors and extract it later during logging or formatting.

### 4. Stack Traces for Root Cause Analysis

When diagnosing production issues, developers need more than just error messages — they need stack traces that show where the error originated. This is especially important when multiple wrapping or rethrowing occurs. By capturing the trace at the point of error creation, this package enables faster debugging and clearer logs.

## Installation

```bash
go get github.com/axkit/errors
```

## Error Template

Predefined errors offer reusable templates for consistent error creation. Use the `Template` function to declare them:

```go
import "github.com/axkit/errors"

var (
	ErrInvalidInput = errors.Template("invalid input provided").
							Code("CRM-0901").
							StatusCode(400).
							Severity(errors.Tiny)
    
	ErrServiceUnavailable = errors.Template("service unavailable").
							Code("SRV-0253").
							StatusCode(500).
							Severity(errors.Critical)

	// Predefined error gets `Tiny` severity  by default.
	ErrInvalidFilter = errors.Template("invalid filtering rule").
							Code("CRM-0042").
							StatusCode(400)

)

if request.Email == "" {
	return ErrInvalidInput.New().Msg("empty email")
}

if request.Age < 18 {
	return ErrInvalidInput.New().Set("age", request.Age).Msg("invalid age")
}

customer, err := service.CustomerByID(request.CustomerID)
if err != nil {
	return ErrServiceUnavailable.Wrap(err)
}

if customer == nil {
	return ErrInvalidInput.New().Msg("invalid customer")
}

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
| `fields`    | Custom key-value pairs for additional context  |
| `stack`     | Stack frames showing the call trace            |

## Capturing the Stack Trace

A stack trace is automatically captured at the moment an error is created or first wrapped. This allows developers to identify where the problem originated, even if the error travels up the call stack.

A stack trace is captured when one of the following methods is called:

- errors.TemplateError.Wrap(...)
- errors.TemplateError.New(...)
- errors.Wrap(...)

> Rewrapping an error does not overwrite an existing stack trace. The original call site remains preserved, ensuring consistent and reliable debugging information.

## Error Logging

Effective error logging is crucial for debugging and monitoring. This package encourages logging errors at the topmost layer of the application, such as an HTTP controller, while lower layers propagate errors with additional context. This ensures that logs are concise and meaningful.

```go
var ErrInvalidObjectID = errors.Template("inalid object id").Code("CRM-0400").StatusCode(400)

// customer_repo.go
customerTable := velum.NewTable[Customer]("customers")

customer, err := customerTable.GetByPK(ctx, db, customerID)
return customer, err

// customer_service.go
customer, err := repo.CustomerByID(customerID)
if err != nil && errors.Is(err, repo.ErrNotFound) {
    return nil, ErrInvalidObjectID.Wrap(err).Set("customerId", customerID)
}

// customer_controller.go
customer, err := service.CustomerByID(customerID)
if err != nil {
	buf := errors.ToJSON(err, errors.WithAttributes(errors.AddStack|errors.AddWrappedErrors))
	
	// server output (extended)
	log.Println(string(buf))
	
	// client output (reduced)
	buf = errors.ToJSON(err)
}
```

### Custom JSON Serialization

If you need to implement a custom JSON serializer, the `errors.Serialize(err)` method provides an object containing all public attributes of the error. This allows you to define your own serialization logic tailored to your application's requirements.

## Alarm Notifications

Set an alarmer to notify on critical errors: 

```go
type CustomAlarmer struct{}

func (c *CustomAlarmer) Alarm(se *SerializedError) {
    fmt.Println("Critical error:", err)
}

errors.SetAlarmer(&CustomAlarmer{})

var ErrConsistencyFailed = errors.Template("data consistency failed").Severity(errors.Critical) 

// CustomAlarmer.Alarm() will be invocated automatically (severity=Critical)
return ErrDataseConnectionFailure.New()
```

## Severity Levels

The package classifies errors into three severity levels:

- **Tiny**: Minor issues, typically validation errors.
- **Medium**: Regular errors that log stack traces.
- **Critical**: Major issues requiring immediate attention.

When wrapping errors, the `severity` and `statusCode` attributes can be overridden. The client will always receive the latest `severity` and `statusCode` values from the outermost error. Any inner errors even with higher severity or different status codes will only be logged, ensuring that the most relevant information is presented to the client while maintaining detailed logs for debugging purposes.

## Migration Guide

Below is a categorized list of how errors are typically created or obtained in Go code. These represent common entry points for error handling.

```go

// 1. plain, unstructured error
return errors.New("message") 

// 2. formatted, unstructured 
return fmt.Errorf("something failed") 

// 3. wrapping + formatting
return fmt.Errorf("wrap: %w", err) 

// 4. Custom struct implementing error
return &CustomError{}

// 5. External packages returning error values
rows, err := db.Query("SELECT COUNT(1) FROM customers")
if err != nil {
	return nil, err 
}

// 6. shared "constants"
var ErrX = errors.New("...")

// 7. Sentinel errors for comparison via errors.Is(...)
// 8. Errors returned from standard library or external packages(e.g., io.EOF, pgx.ErrNoRows) 
```

### Migration Strategy

Migrating to `axkit/errors` can be done incrementally, without needing to rewrite your entire codebase at once. The primary goal is to transition from unstructured error handling to a system of reusable, structured templates with support for metadata, stack traces, and observability.

#### Step 1: Replace the Import

Start by replacing the standard library import:

```go
import "errors"
````

with:

```go
import "github.com/axkit/errors"
```

This change is non-breaking: errors.New(...)  remains available and behaves the same way, returning a basic error without stack trace. This allows your application to compile and function as before.


#### Step 2: Identify and Replace Static Errors with Templates

Begin replacing predefined or shared errors like:

```go
var ErrUnauthorized = errors.New("unauthorized")
```

with structured templates:

```go
var ErrUnauthorized = errors.Template("unauthorized").
	Code("AAA-0401").
	StatusCode(401).
	Severity(Tiny)
```

I recommend placing templates at the top of each file or organizing them into a dedicated file such as errors.go, error_templates.go, etc.

#### Step 3: Wrap External or Lower-Level Errors

Whenever you receive an error from the standard library or a third-party package (e.g. pgx.ErrNoRows, io.EOF, sql.ErrTxDone), wrap it in your own context using a template:

```go
customers, err := s.repo.CustomerByID(customerID)
if err == pgx.ErrNoRows {
	return ErrCustomerNotFound.Wrap(err)
}
```

If no reusable template exists yet, it’s acceptable to inline one during early migration:

```go
customers, err := s.repo.CustomerByID(customerID)
if err == pgx.ErrNoRows {
	return errors.Wrap(err, "customer not found").StatusCode(400).Set("customerId", customerID)
}
```

### Step 4: Centralize and Document Templates

As migration progresses, ensure that all templates:

- Have a unique Code(...) identifier
- Are grouped and reusable
- Are linked to documentation or support systems (e.g. HelpDesk, monitoring, alerting)
- Capture stack trace at the appropriate level via .New() or .Wrap()

This ensures a consistent and observable error-handling experience across your application.

#### Summary

| Step    | Goal                                | Complexity         | Backward Compatible |
|---------|-------------------------------------|--------------------|----------------------|
| Step 1 | Replace standard import              | Very Low           | ✅ Yes               |
| Step 2 | Use `Template` for shared errors     | Medium             | ✅ Yes               |
| Step 3 | Wrap external or third-party errors  | Medium             | ✅ Yes               |
| Step 4 | Centralize and document templates    | High (but worth it)| ✅ Yes               |

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
