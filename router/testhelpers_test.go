package router_test

import wspulse "github.com/wspulse/core"

type mockConnection struct {
	id     string
	roomID string
}

func newMockConnection(id, roomID string) *mockConnection {
	return &mockConnection{id: id, roomID: roomID}
}

func (m *mockConnection) ID() string { return m.id }

func (m *mockConnection) RoomID() string { return m.roomID }

func (m *mockConnection) Send(_ wspulse.Frame) error { return nil }

func (m *mockConnection) Close() error { return nil }

func (m *mockConnection) Done() <-chan struct{} {
	ch := make(chan struct{})
	return ch
}
