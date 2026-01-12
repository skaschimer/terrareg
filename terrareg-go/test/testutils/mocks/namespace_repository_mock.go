package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// MockNamespaceRepository is a mock for NamespaceRepository
type MockNamespaceRepository struct {
	mock.Mock
}

// Ensure MockNamespaceRepository implements the interface at compile time
var _ repository.NamespaceRepository = (*MockNamespaceRepository)(nil)

// Save mocks saving a namespace
func (m *MockNamespaceRepository) Save(ctx context.Context, namespace *model.Namespace) error {
	args := m.Called(ctx, namespace)
	// Set ID on the namespace if provided
	if id, ok := args.Get(0).(int); ok && id > 0 {
		// Note: Namespace model doesn't expose a direct ID setter, so we'll rely on the test to handle this
		_ = id
	}
	return args.Error(1)
}

// FindByID mocks finding a namespace by ID
func (m *MockNamespaceRepository) FindByID(ctx context.Context, id int) (*model.Namespace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Namespace), args.Error(1)
}

// FindByName mocks finding a namespace by name
func (m *MockNamespaceRepository) FindByName(ctx context.Context, name string) (*model.Namespace, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Namespace), args.Error(1)
}

// List mocks listing all namespaces
func (m *MockNamespaceRepository) List(ctx context.Context) ([]*model.Namespace, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return []*model.Namespace{}, args.Error(1)
	}
	return args.Get(0).([]*model.Namespace), args.Error(1)
}

// Delete mocks deleting a namespace
func (m *MockNamespaceRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Exists mocks checking if a namespace exists
func (m *MockNamespaceRepository) Exists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}
