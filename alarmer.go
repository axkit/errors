package errors

// Alarmer is an interface wrapping a single method Alarm
//
// Alarm is invocated automatically when critical error is caught,
// if alarmer is set.
type Alarmer interface {
	Alarm(err error)
}

var alarmer Alarmer

// SetAlarmer sets Alarmer implementation to be used when critical error is caught.
func SetAlarmer(a Alarmer) {
	alarmer = a
}
