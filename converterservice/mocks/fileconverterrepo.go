// Mocks FileConverterRepository
package mocks

import (
	"errors"
	"fmt"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice/db"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice/enums"
	"time"
)

type MockFileConverterRepo struct {
	Data    map[string]*db.ConvertJob
	Success bool
}

func NewMockFileConverterRepo() *MockFileConverterRepo {
	return &MockFileConverterRepo{
		Data:    make(map[string]*db.ConvertJob),
		Success: true,
	}
}

func (m *MockFileConverterRepo) NewRequest(id string) (bool, error) {
	if m.Success {
		m.Data[id] = &db.ConvertJob{
			Id: id,
			Status: enums.QUEUED.Name(),
			CurrUrl: "NONE",
			LastUpdated: time.Now(),
		}
		return true, nil
	}
	return false, errors.New(fmt.Sprintf("failed to create request %s", id))
}

func (m *MockFileConverterRepo) StartConversion(id string) (bool, error) {
	if m.Success && m.Data[id] != nil {
		job := m.Data[id]
		job.Status = enums.CONVERTING.Name()
		job.LastUpdated = time.Now()
		return true, nil
	}
	return false, errors.New(fmt.Sprintf("failed to set status to converting in DB for id %s", id))
}

func (m *MockFileConverterRepo) CompleteConversion(id string, url string) (bool, error) {
	if m.Success && m.Data[id] != nil {
		job := m.Data[id]
		job.CurrUrl = url
		job.Status = enums.COMPLETED.Name()
		job.LastUpdated = time.Now()
		return true, nil
	}
	return false, errors.New(fmt.Sprintf("failed to set completion in DB for id %s and url %s", id, url))
}

func (m *MockFileConverterRepo) FailConversion(id string) (bool, error) {
	if m.Success && m.Data[id] != nil {
		job := m.Data[id]
		job.Status = enums.FAILED.Name()
		job.LastUpdated = time.Now()
		return true, nil
	}
	return false, errors.New(fmt.Sprintf("failed to set failure in DB for Id %s", id))
}

func (m *MockFileConverterRepo) GetConversion(id string) (*db.ConvertJob, error) {
	if m.Success && m.Data[id] != nil {
		return m.Data[id], nil
	}
	return nil, errors.New(fmt.Sprintf("could not get job by id %s", id))
}
