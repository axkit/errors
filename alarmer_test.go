package errors

import (
	"errors"
	"testing"
)

type MockAlarmer struct {
	called bool
	err    error
}

func (m *MockAlarmer) Alarm(err error) {
	m.called = true
	m.err = err
}

func TestSetAlarmer(t *testing.T) {
	mock := &MockAlarmer{}
	SetAlarmer(mock)

	if alarmer != mock {
		t.Errorf("Expected alarmer to be set to mock, but got %v", alarmer)
	}
}

func TestAlarm(t *testing.T) {
	mock := &MockAlarmer{}
	SetAlarmer(mock)

	testErr := errors.New("test error")
	alarmer.Alarm(testErr)

	if !mock.called {
		t.Errorf("Expected Alarm to be called")
	}

	if mock.err != testErr {
		t.Errorf("Expected error to be %v, but got %v", testErr, mock.err)
	}
}
