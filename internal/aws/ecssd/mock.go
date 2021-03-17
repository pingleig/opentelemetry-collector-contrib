package ecssd

import (
	"context"
)

// MockFecther allows running ServiceDiscovery without using actual TaskFetcher.
type MockFetcher struct {
	// use a factory instead of static tasks because filter and exporter
	// will udpate tasks in place. This avoid introducing a deep copy library for unt test.
	factory func() []*Task
}

func newMockFetcher(tasksFactory func() []*Task) *MockFetcher {
	return &MockFetcher{factory: tasksFactory}
}

// FetchAndDecorate calls factory to create a new list of task everytime.
func (m *MockFetcher) FetchAndDecorate(_ context.Context) ([]*Task, error) {
	return m.factory(), nil
}
