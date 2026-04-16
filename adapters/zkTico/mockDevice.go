package main


type MockZKDevice struct {
	Events    chan *AttendanceEvent
	Connected bool
	Stopped   bool
}

func (m *MockZKDevice) Connect() error {
	m.Connected = true
	return nil
}

func (m *MockZKDevice) LiveCapture() (<-chan *AttendanceEvent, error) {
	return m.Events, nil
}

func (m *MockZKDevice) StopCapture() {
	m.Stopped = true
	close(m.Events)
}

func NewMockZKDevice() *MockZKDevice {
	return &MockZKDevice{
		Events: make(chan *AttendanceEvent, 10),
	}
}