package mocks

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice/fileconverter"
	"io"
	"os"
)

type MockExecutableFactory struct {
	Success bool
	Data    map[string]*MockExecutable
}

type MockExecutable struct {
	Success bool
	Job     *fileconverter.ConversionAttributes
}

func NewMockExecutableFactory() *MockExecutableFactory {
	return &MockExecutableFactory{
		Success: true,
		Data: make(map[string]*MockExecutable),
	}
}

func (m *MockExecutableFactory) SelectCommand(job *fileconverter.ConversionAttributes) fileconverter.Executable {
	executable := &MockExecutable{
		Success: m.Success,
		Job: job,
	}
	job.TmpFile = fmt.Sprintf("/tmp/%s", job.Request.Id)
	m.Data[job.Request.Id] = executable
	return executable
}

func (m *MockExecutable) Start() error {
	if m.Success {
		_, err := os.Create(m.Job.TmpFile)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("command failed to execute")
}

func (m *MockExecutable) Wait() error {
	if m.Success {
		return nil
	}
	return errors.New("error encountered during wait")
}

func (m *MockExecutable) Stdout() io.Writer {
	return bytes.NewBuffer(make([]byte, 1024))
}

func (m *MockExecutable) SetStdout(stdout io.Writer) {
}

func (m *MockExecutable) Stdin() io.Reader {
	return bytes.NewReader(make([]byte, 1024))
}

func (m *MockExecutable) SetStdin(stdin io.Reader) {
}

func (m *MockExecutable) Stderr() io.Writer {
	return bytes.NewBuffer(make([]byte, 1024))
}

func (m *MockExecutable) SetStderr(stderr io.Writer) {
}
