package errors_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/axkit/errors"
	"github.com/rs/zerolog"
)

func f() {
	e := errors.New("invalid rules")
	fmt.Println("e.Error()=", e.Error())

	es := errors.Capture(e).Msg("end of file reached").Set("file", "config.json").SetSeverity(500)
	fmt.Println("es.Error()=", es.Error())

	js := errors.Capture(es).Msg("file reading error")
	fmt.Println("js.Error()=", js.Error())

	arr := js.Children()
	for i := range arr {
		fmt.Printf("%d = %#v\n", i, arr[i])
	}
	fmt.Println("")
	fmt.Printf("e=%#v\n", e)

}
func TestCapture(t *testing.T) {

	f()
	dbe := errors.CaptureDatabaseError(io.ErrNoProgress)
	fmt.Println("")
	fmt.Printf("dbe=%#v\n", dbe)

}

func TestMarshalZerologObject(t *testing.T) {

	l := zerolog.New(os.Stdout)
	e := errors.New("invalid rules")
	//fmt.Println("e.Error()=", e.Error())
	es := errors.Capture(e).Msg("end of file reached").Set("file", "config.json").SetStatusCode(500).SetSeverity(errors.Critical)
	//fmt.Println("es.Error()=", es.Error())

	es.Log(&l)
}
