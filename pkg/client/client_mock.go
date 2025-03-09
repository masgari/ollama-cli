package client

import (
	"context"

	"github.com/ollama/ollama/api"
	"github.com/stretchr/testify/mock"
)

// MockClientTestify is a mock implementation of the Client interface using testify/mock
type MockClientTestify struct {
	mock.Mock
}

// ListModels implements the Client interface
func (m *MockClientTestify) ListModels(ctx context.Context) (*api.ListResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.ListResponse), args.Error(1)
}

// GetModelDetails implements the Client interface
func (m *MockClientTestify) GetModelDetails(ctx context.Context, modelName string) (*api.ShowResponse, error) {
	args := m.Called(ctx, modelName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.ShowResponse), args.Error(1)
}

// DeleteModel implements the Client interface
func (m *MockClientTestify) DeleteModel(ctx context.Context, modelName string) error {
	args := m.Called(ctx, modelName)
	return args.Error(0)
}

// PullModel implements the Client interface
func (m *MockClientTestify) PullModel(ctx context.Context, modelName string) error {
	args := m.Called(ctx, modelName)
	return args.Error(0)
}

// NewMockClient creates a new testify mock client
func NewMockClient() *MockClientTestify {
	return &MockClientTestify{}
}
