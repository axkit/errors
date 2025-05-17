package errors

import "bytes"

// SeverityLevel describes error severity levels.
type SeverityLevel int

const (
	// Tiny classifies expected, managed errors that do not require administrator attention.
	// Writing a call stack to the journal file is not recommended.
	//
	// Example: error related to validation of entered form fields.
	Tiny SeverityLevel = iota

	// Medium classifies a regular error. A call stack is written to the log.
	Medium

	// Critical classifies a significant error, requiring immediate attention.
	// The occurrence of the error should be communicated to the administrator
	// through all available channels. A call stack is written to the log.
	// If Alarmer is set, it will be called.
	Critical
)

var (
	// Important: quotes are included.
	tiny     = []byte(`"tiny"`)
	medium   = []byte(`"medium"`)
	critical = []byte(`"critical"`)
	unknown  = []byte(`"unknown"`)
)

const (
	stiny     = "tiny"
	smedium   = "medium"
	scritical = "critical"
	sunknown  = "unknown"
)

// String returns severity level string representation.
func (sl SeverityLevel) String() string {
	switch sl {
	case Tiny:
		return stiny
	case Medium:
		return smedium
	case Critical:
		return scritical
	}
	return sunknown
}

// MarshalJSON implements json/Marshaller interface.
func (sl SeverityLevel) MarshalJSON() ([]byte, error) {
	switch sl {
	case Tiny:
		return tiny, nil
	case Medium:
		return medium, nil
	case Critical:
		return critical, nil
	}
	return unknown, nil
}

// UnmarshalJSON implements json/Unmarshaller interface.
func (sl *SeverityLevel) UnmarshalJSON(data []byte) error {
	switch {
	case bytes.Equal(data, tiny):
		*sl = Tiny
	case bytes.Equal(data, medium):
		*sl = Medium
	case bytes.Equal(data, critical):
		*sl = Critical
	default:
		*sl = 0
	}
	return nil
}
