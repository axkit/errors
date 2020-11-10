# errors [![GoDoc](https://godoc.org/github.com/axkit/errors?status.svg)](https://godoc.org/github.com/axkit/errors) [![Build Status](https://travis-ci.org/axkit/errors.svg?branch=master)](https://travis-ci.org/axkit/errors) [![Coverage Status](https://coveralls.io/repos/github/errors/gonfig/badge.svg)](https://coveralls.io/github/axkit/errors) [![Go Report Card](https://goreportcard.com/badge/github.com/axkit/errors)](https://goreportcard.com/report/github.com/axkit/errors)

Enterprise approach of error handling

# Goals

- Capture stack at once in the begining.
- Enhance error with key/value pairs, later to be written into structurized log.
- Enhance error with the code, what can be refered in documentation. (i.g. ORA-0600 in Oracle).
- Enhance error with severity level.
- Ability of drill down to the wrapped errors.
- Chaining.

## Usage

```
import (
    "github.com/axkit/errors"
)

func WriteJSON(w io.Writer, src interface{}) error {

    buf, err := json.Marshal(src)
    if err != nil {
        return errors.Catch(err).Critical().Set("obj", src).StatusCode(500)
    }
    return nil
}


func Div(a, b int) (int, error) {

    if b == 0 {
        return 0, errors.New("divizion by zero").Critica().SetPairs("a", a)
    }

    return a/b, nil
}

func (srv *CustomerService)CustomerByID(id int) (*Customer, error) {

    c, ok := srv.repo.CustomerByID(id)
    if !ok {
        return nil, errors.New("customer not found").Critical().Set("id", id).StatusCode(404)
    }

    return c, nil
}


```

Prague 2020
