// Mocks FileConverterData
package mocks

import (
	"errors"
	"fmt"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice/enums"
	"time"
)

type MockFileConverterDb struct {
	Data    map[string]*converterservice.ConvertJob
	Success bool
}

func NewMockFileConverterDb() *MockFileConverterDb {
	return &MockFileConverterDb{
		Data:    make(map[string]*converterservice.ConvertJob),
		Success: true,
	}
}

func (m *MockFileConverterDb) NewRequest(id string) (bool, error) {
	if m.Success {
		m.Data[id] = &converterservice.ConvertJob{
			Id: id,
			Status: enums.QUEUED.Name(),
			CurrUrl: "NONE",
			LastUpdated: time.Now(),
		}
		return true, nil
	}
	return false, errors.New(fmt.Sprintf("failed to create request %s", id))
}

func (m *MockFileConverterDb) StartConversion(id string) (bool, error) {
	if m.Success && m.Data[id] != nil {
		job := m.Data[id]
		job.Status = enums.CONVERTING.Name()
		job.LastUpdated = time.Now()
		return true, nil
	}
	return false, errors.New(fmt.Sprintf("failed to set status to converting in DB for id %s", id))
}

func (m *MockFileConverterDb) CompleteConversion(id string, url string) (bool, error) {
	if m.Success && m.Data[id] != nil {
		job := m.Data[id]
		job.CurrUrl = url
		job.Status = enums.COMPLETED.Name()
		job.LastUpdated = time.Now()
		return true, nil
	}
	return false, errors.New(fmt.Sprintf("failed to set completion in DB for id %s and url %s", id, url))
}

func (m *MockFileConverterDb) FailConversion(id string) (bool, error) {
	if m.Success && m.Data[id] != nil {
		job := m.Data[id]
		job.Status = enums.FAILED.Name()
		job.LastUpdated = time.Now()
		return true, nil
	}
	return false, errors.New(fmt.Sprintf("failed to set failure in DB for Id %s", id))
}

func (m *MockFileConverterDb) GetConversion(id string) (*converterservice.ConvertJob, error) {
	if m.Success && m.Data[id] != nil {
		return m.Data[id], nil
	}
	return nil, errors.New(fmt.Sprintf("could not get job by id %s", id))
}
