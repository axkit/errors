package errors

type Alarmer interface {
	Alarm(err error)
}

var alarmer Alarmer

func SetAlarmer(a Alarmer) {
	alarmer = a
}
