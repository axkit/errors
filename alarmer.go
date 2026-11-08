package errors

// Alarmer is an interface wrapping a single method Alarm.
//
// Alarm is called by Error.Alarm and when ToJSON fails to marshal an error,
// if alarmer is set.
type Alarmer interface {
	Alarm(err error)
}

var alarmer Alarmer

// SetAlarmer sets Alarmer implementation to be used by Error.Alarm and ToJSON.
func SetAlarmer(a Alarmer) {
	alarmer = a
}
