# errors [![GoDoc](https://pkg.go.dev/badge/github.com/errors/errors?status.svg)](https://pkg.go.dev/github.com/axkit/errors) [![Build Status](https://travis-ci.org/axkit/errors.svg?branch=main)](https://travis-ci.org/axkit/errors) [![Coverage Status](https://coveralls.io/repos/github/axkit/errors/badge.svg)](https://coveralls.io/github/akkit/errors) [![Go Report Card](https://goreportcard.com/badge/github.com/axkit/errors)](https://goreportcard.com/report/github.com/axkit/errors)

# errors
The errors package provides an enterprise approach of error handling. Drop in replacement of standard errors package.

## Motivation
Make errors helpful for quick problem localization. Reduce amount of emails to the helpdesk due to better explained error reason. 

# Requirements

- Wrapping: enhance original error message with context specific one;
- Capture the calling stack;
- Cut calling stack related to the HTTP framework;
- Enhance error with key/value pairs, later could be written into structurized log;
- Enhance error with the code what can be refered in documentation. (i.g. ORA-0600 in Oracle);
- Enhance error with severity level;
- Support different JSON representation for server and client; 
- Possibility to mark any error as protected. It will not be presented in client's JSON.
- Notify SRE if needed.

## Installation
```
go get -u github.com/axkit/error
```

## Usage Examples

### Catch and Enhance Standard Go Error 
```
func (srv *CustomerService)WriteJSON(w io.Writer, c *Customer) (int, error) {

    buf, err := json.Marshal(src)
    if err != nil {
        return 0, errors.Catch(err).Critical().Set("customer", c).StatusCode(500).Msg("internal error")
    }

    n, err := w.Write(buf)
    if err != nil {
        // Level is Tiny by default. 
        return 0, errors.Catch(err).StatusCode(500).Msg("writing to stream failed").Code("APP-0001")
    }
     
    return n, nil  
}

```

### Catch and Enhance Already Catched Error 
```
func AllowedProductAmount(balance, price int) (int, error) {

    res, err := Calc(balance, price)
    if err != nil {
        return 0, errors.Catch(err).SetPairs("balance", balance, "price", price).Msg("no allowed products")
    }

    return res, nil
}


func Calc(a, b int) (int, error) {

    if b == 0 {
        return 0, errors.New("divizion by zero").Critical()
    }

    return a/b, nil
}
```

### Recatch NotFound Error Conditionally
There is a special function ```errors.IsNotFound()``` that returns true error has StatusCode = 404 or created using ```errors.NotFound()```.
```
func (srv *CustomerService)AcceptPayment(customerID int, paymentAmount int64) error {

    c, err := srv.repo.CustomerByID(id)
    if err != nil {
        if errors.IsNotFound(err) {
            return nil, errors.Catch(err).Medium().Msg("invalid customer")
        }
        return return nil, errors.Catch(err).Critical().Msg("internal error")
    }

    return c, nil
}

func (srv *CustomerService)CustomerByID(id int) (*Customer, error) {

    c, ok := srv.repo.CustomerByID(id)
    if !ok {
        return nil, errors.NotFound("customer not found").Set("id", id)
    }

    return c, nil
}
```
### Write Error Responce to Client
```
    err := doSmth()
    
    // err is standard error  
    fmt.Println(errors.ToClientJSON(err))

    // Output:
    {"msg":"result of Error() method"}

```

### Write Error Responce to Server Log


### Send Alarm to SRE


## License
MIT

