// Code generated by MockGen. DO NOT EDIT.
// Source: cloudpix/internal/domain/repository (interfaces: UploadMetadataRepository)
//
// Generated by this command:
//
//	mockgen -destination=internal/mocks/repository/mock_upload_metadata_repository.go -package=repository cloudpix/internal/domain/repository UploadMetadataRepository
//

// Package repository is a generated GoMock package.
package repository

import (
	model "cloudpix/internal/domain/model"
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockUploadMetadataRepository is a mock of UploadMetadataRepository interface.
type MockUploadMetadataRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUploadMetadataRepositoryMockRecorder
	isgomock struct{}
}

// MockUploadMetadataRepositoryMockRecorder is the mock recorder for MockUploadMetadataRepository.
type MockUploadMetadataRepositoryMockRecorder struct {
	mock *MockUploadMetadataRepository
}

// NewMockUploadMetadataRepository creates a new mock instance.
func NewMockUploadMetadataRepository(ctrl *gomock.Controller) *MockUploadMetadataRepository {
	mock := &MockUploadMetadataRepository{ctrl: ctrl}
	mock.recorder = &MockUploadMetadataRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUploadMetadataRepository) EXPECT() *MockUploadMetadataRepositoryMockRecorder {
	return m.recorder
}

// SaveMetadata mocks base method.
func (m *MockUploadMetadataRepository) SaveMetadata(ctx context.Context, metadata *model.UploadMetadata) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveMetadata", ctx, metadata)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveMetadata indicates an expected call of SaveMetadata.
func (mr *MockUploadMetadataRepositoryMockRecorder) SaveMetadata(ctx, metadata any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveMetadata", reflect.TypeOf((*MockUploadMetadataRepository)(nil).SaveMetadata), ctx, metadata)
}
