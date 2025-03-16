// Code generated by MockGen. DO NOT EDIT.
// Source: cloudpix/internal/domain/repository (interfaces: TagRepository)
//
// Generated by this command:
//
//	mockgen -destination=internal/mocks/repository/mock_tag_repository.go -package=mocks cloudpix/internal/domain/repository TagRepository
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockTagRepository is a mock of TagRepository interface.
type MockTagRepository struct {
	ctrl     *gomock.Controller
	recorder *MockTagRepositoryMockRecorder
	isgomock struct{}
}

// MockTagRepositoryMockRecorder is the mock recorder for MockTagRepository.
type MockTagRepositoryMockRecorder struct {
	mock *MockTagRepository
}

// NewMockTagRepository creates a new mock instance.
func NewMockTagRepository(ctrl *gomock.Controller) *MockTagRepository {
	mock := &MockTagRepository{ctrl: ctrl}
	mock.recorder = &MockTagRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTagRepository) EXPECT() *MockTagRepositoryMockRecorder {
	return m.recorder
}

// AddTags mocks base method.
func (m *MockTagRepository) AddTags(ctx context.Context, imageID string, tags []string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddTags", ctx, imageID, tags)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddTags indicates an expected call of AddTags.
func (mr *MockTagRepositoryMockRecorder) AddTags(ctx, imageID, tags any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddTags", reflect.TypeOf((*MockTagRepository)(nil).AddTags), ctx, imageID, tags)
}

// GetImageTags mocks base method.
func (m *MockTagRepository) GetImageTags(ctx context.Context, imageID string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetImageTags", ctx, imageID)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetImageTags indicates an expected call of GetImageTags.
func (mr *MockTagRepositoryMockRecorder) GetImageTags(ctx, imageID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetImageTags", reflect.TypeOf((*MockTagRepository)(nil).GetImageTags), ctx, imageID)
}

// ListTags mocks base method.
func (m *MockTagRepository) ListTags(ctx context.Context) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListTags", ctx)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListTags indicates an expected call of ListTags.
func (mr *MockTagRepositoryMockRecorder) ListTags(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListTags", reflect.TypeOf((*MockTagRepository)(nil).ListTags), ctx)
}

// RemoveAllTags mocks base method.
func (m *MockTagRepository) RemoveAllTags(ctx context.Context, imageID string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveAllTags", ctx, imageID)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RemoveAllTags indicates an expected call of RemoveAllTags.
func (mr *MockTagRepositoryMockRecorder) RemoveAllTags(ctx, imageID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveAllTags", reflect.TypeOf((*MockTagRepository)(nil).RemoveAllTags), ctx, imageID)
}

// RemoveTags mocks base method.
func (m *MockTagRepository) RemoveTags(ctx context.Context, imageID string, tags []string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveTags", ctx, imageID, tags)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RemoveTags indicates an expected call of RemoveTags.
func (mr *MockTagRepositoryMockRecorder) RemoveTags(ctx, imageID, tags any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveTags", reflect.TypeOf((*MockTagRepository)(nil).RemoveTags), ctx, imageID, tags)
}

// VerifyImageExists mocks base method.
func (m *MockTagRepository) VerifyImageExists(ctx context.Context, imageID string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyImageExists", ctx, imageID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// VerifyImageExists indicates an expected call of VerifyImageExists.
func (mr *MockTagRepositoryMockRecorder) VerifyImageExists(ctx, imageID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyImageExists", reflect.TypeOf((*MockTagRepository)(nil).VerifyImageExists), ctx, imageID)
}
