package router_test

import (
	"sync"

	wspulse "github.com/wspulse/core"
)

type mockConnection struct {
	id     string
	roomID string
	done   chan struct{}
	once   sync.Once
}

func newMockConnection(id, roomID string) *mockConnection {
	return &mockConnection{id: id, roomID: roomID, done: make(chan struct{})}
}

func (m *mockConnection) ID() string { return m.id }

func (m *mockConnection) RoomID() string { return m.roomID }

func (m *mockConnection) Send(_ wspulse.Message) error { return nil }

func (m *mockConnection) Close() error {
	m.once.Do(func() { close(m.done) })
	return nil
}

func (m *mockConnection) Done() <-chan struct{} {
	return m.done
}
