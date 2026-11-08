---
name: axkit-errors
description: Idiomatic error handling with github.com/axkit/errors in Go. Use when writing, reviewing or refactoring Go code that imports github.com/axkit/errors, or when the project uses it and you declare error templates, create or wrap errors, check them with Is/As, log them or return them as JSON from an HTTP handler.
---

# axkit/errors

`github.com/axkit/errors` adds structured errors to Go: reusable templates with a code, HTTP status and severity, custom fields, stack traces, and JSON output with different detail levels for the server log and the client.

## Model

- **Template** (`*errors.ErrorTemplate`): a package-level declaration of an error kind. Created once, never changed at runtime.
- **Error** (`*errors.Error`): an instance created from a template by `New()` or `Wrap(err)`. Carries fields and the stack trace of the place where it was created.
- **Chain**: every wrap keeps the previous error. `Error()` returns all messages joined by `": "`; JSON writes each level separately.
- **Output**: errors are logged once, at the top (HTTP handler, job runner), with `ToJSON`.

## Rules

### 1. Import it instead of the standard package

```go
import "github.com/axkit/errors"
```

The package provides the whole standard API (`New`, `Is`, `As`, `Unwrap`, `Join`, `ErrUnsupported`, `AsType`). Do not import both; do not alias the standard package.

### 2. Declare templates at package level

```go
var (
	ErrCustomerNotFound = errors.Template("customer not found").
		Code("CRM-0404").StatusCode(404).Severity(errors.Tiny)

	ErrInvalidEmail = errors.Template("invalid email").
		Code("CRM-0400").StatusCode(400).Severity(errors.Tiny)

	ErrDatabase = errors.Template("database failure").
		Code("CRM-0500").StatusCode(500).Severity(errors.Medium).Protected(true)
)
```

- Give every template a unique `Code`: it is the stable reference for support and documentation. Look at existing codes in the project and follow their prefix and numbering.
- Set `StatusCode` for errors that reach an HTTP response.
- Set `Severity`: `Tiny` for expected errors (validation, not found), `Medium` for failures, `Critical` for errors that need immediate attention. Without it the severity is `Unknown`.
- Mark templates whose message reveals internals (SQL, hosts, file paths) with `Protected(true)`.
- Template setters (`Code`, `Set`, `StatusCode`, ...) modify the shared template. Call them only in the declaration, never in request code: it is a data race and changes the error for everyone.
- Never create templates inside functions.

### 3. Create an error where the problem is detected

```go
if !isValidEmail(req.Email) {
	return ErrInvalidEmail.New().Set("email", req.Email)
}
```

- `New()` captures the stack trace at this line.
- Put context into fields with `Set(key, value)`. Values must be JSON-serializable. Do not put passwords, tokens or full connection strings into fields: they go to logs.
- `Msg("...")` replaces the message, e.g. to make it more specific for the client. The error still matches its template in `errors.Is`.

### 4. Wrap errors from other packages where they enter your code

```go
row, err := db.QueryContext(ctx, q, id)
if err != nil {
	return ErrDatabase.Wrap(err).Set("customerId", id)
}
```

- Wrap third-party and standard errors (`sql.ErrNoRows`, `io.EOF`, `*pgconn.PgError`) with a template at the boundary. Do not return them bare.
- To translate an error into another kind, check it and wrap with the matching template:

```go
if errors.Is(err, sql.ErrNoRows) {
	return ErrCustomerNotFound.Wrap(err).Set("customerId", id)
}
```

### 5. Add context in upper layers with errors.Wrap

```go
customer, err := s.repo.CustomerByID(ctx, id)
if err != nil {
	return nil, errors.Wrap(err, "load customer for invoice")
}
```

- `errors.Wrap(err, msg)` keeps code, status code, severity, protected flag, fields and stack of `err`; it only adds a message level.
- The message of the outermost error is the `msg` the client gets. Write `errors.Wrap` messages that are fine to show to the client, or return the error as is.
- Use `OtherTemplate.Wrap(err)` when the layer changes the meaning of the error, e.g. a protected database error becomes a public "cannot create invoice".
- Prefer `errors.Wrap` to `fmt.Errorf("...: %w", err)`: `fmt.Errorf` works with `Is`/`As`, but its message is written to JSON as one concatenated string.
- Return the error as is (`return err`) when there is nothing to add. Do not wrap at every layer by habit.
- The stack is captured once, where the error was created. Wrapping does not replace it.

### 6. Check errors with Is, extract them with AsType or As

```go
if errors.Is(err, ErrCustomerNotFound) { ... }   // compare with the template

if e, ok := errors.AsType[*errors.Error](err); ok { ... }   // get the first *errors.Error

var pgErr *pgconn.PgError
if errors.As(err, &pgErr) { ... }                 // get a third-party error from the chain
```

- Compare with the template variable, not with `err.Error()` strings and not with a new instance.
- An error matches its template after `Msg`, `StatusCode` and other instance setters.
- For `As` with `*errors.Error`, pass a pointer to a nil `*errors.Error` (or use `AsType`). Do not pass a non-nil instance as a filter; use `Is` to check the kind.

### 7. Protect internal details

- A protected error must not be the top-level error returned to the client: its message is written to JSON as is. Wrap it with a public template:

```go
return ErrCreateInvoice.Wrap(err)   // err may contain a protected ErrDatabase
```

- `errors.Wrap(err, msg)` keeps the protected flag, it does not make the error public.
- Protected errors in the chain are written to JSON only with `AddProtected` (included in `ServerOutputFormat`).

### 8. Log once at the top, respond with the reduced JSON

```go
// at startup
clientFormat := errors.ClientOutputFormat
if cfg.Env == "dev" {
	clientFormat = errors.ClientDebugOutputFormat // same JSON as the server log
}

// in the HTTP layer
func writeError(w http.ResponseWriter, err error, clientFormat errors.ErrorSerializationRule) {
	log.Println(string(errors.ToJSON(err, errors.WithAttributes(errors.ServerOutputFormat))))

	status := errors.Serialize(err).StatusCode
	if status == 0 {
		status = http.StatusInternalServerError
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(errors.ToJSON(err, errors.WithAttributes(clientFormat)))
}
```

- Lower layers return errors, they do not log them. The top layer logs each error once.
- `ToJSON` and `Serialize` without options write only message, severity, code and status code. Fields, stack and wrapped errors are added by `WithAttributes(...)`:

| Rule | Adds |
|---|---|
| `AddFields` | `fields` set by `Set` |
| `AddStack` | `stack`, frames as `{"func","file":"path:line"}` |
| `AddWrappedErrors` | `wrapped`: errors below the top-level one |
| `AddProtected` | protected errors in `wrapped` |
| `IndentJSON` | indented output |

- `ServerOutputFormat` includes everything; `ClientOutputFormat` includes nothing extra.
- `WithRootLevelFields([]string{"requestId"})` moves the listed fields to the JSON root; they are written even without `AddFields`.
- `WithStopStackOn("net/http")` cuts the stack at the first frame of a library you are not interested in.
- The package has no getters for code or status code; read them from `errors.Serialize(err)`.

### 9. Alarms for critical errors

```go
errors.SetAlarmer(myAlarmer) // at startup; Alarm(err error)

err := ErrLedgerMismatch.New().Set("accountId", id)
err.Alarm() // explicit, nothing is called automatically
return err
```

## Anti-patterns

| Do not | Do |
|---|---|
| `errors.New("customer not found")` for a domain error | a package-level `errors.Template(...)` with a code |
| `return err` from `database/sql` up to the handler | `ErrDatabase.Wrap(err)` in the repository |
| `ErrX.Set("id", id)` in request code | `ErrX.New().Set("id", id)` |
| `log.Println(err); return err` in a service | `return errors.Wrap(err, "...")`, log in the handler |
| `strings.Contains(err.Error(), "not found")` | `errors.Is(err, ErrCustomerNotFound)` |
| `errors.As(err, &someInstance)` to check the kind | `errors.Is(err, ErrX)` |
| `ToJSON(err, WithAttributes(ServerOutputFormat))` in the client response in production | `ClientOutputFormat` |
| returning a protected error to the client unwrapped | `PublicTemplate.Wrap(err)` |
| `Set("dsn", connString)` with credentials | `Set("dbHost", host)` |

## Tests

```go
err := svc.CreateInvoice(ctx, req)
if !errors.Is(err, ErrCustomerNotFound) {
	t.Fatalf("expected ErrCustomerNotFound, got %v", err)
}
```

Check the kind with `errors.Is`, not the message text.
