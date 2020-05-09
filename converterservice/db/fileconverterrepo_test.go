// Test suite for the file converter repo
package db

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice/enums"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	errorExpectedError = errors.New("expected error but none was received")
	testingError = errors.New("testing error")
)

type AnyTime struct {}

type testingBasics struct {
	id   string
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo FileConverterRepository
}

// returns true when the time is within 1 second
func (a AnyTime) Match(v driver.Value) bool {
	valTime := v.(time.Time)
	diff := time.Now().Sub(valTime)
	return diff <= time.Second
}

func BeforeEach(t *testing.T) *testingBasics {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Error(err.Error())
	}
	repo := NewFromConnection(db)
	id := uuid.New().String()
	return &testingBasics{
		id: id,
		db: db,
		mock: mock,
		repo: repo,
	}
}

func AfterEach(t *testing.T, b *testingBasics) {
	if err := b.mock.ExpectationsWereMet(); err != nil {
		t.Error(err.Error())
	}
	b.db.Close()
}

func TestNewWithConn(t *testing.T) {
	db, _, err := sqlmock.New()
	defer db.Close()
	assert.Nil(t, err)
	repo := NewFromConnection(db)
	assert.NotNil(t, repo)
}

func TestFileConverterData_NewRequest(t *testing.T) {
	b := BeforeEach(t)
	defer AfterEach(t, b)
	b.mock.ExpectExec(fmt.Sprintf("INSERT INTO %s", tableName)).
		WithArgs(b.id, enums.QUEUED.Name(), "NONE", AnyTime{}).
		WillReturnResult(sqlmock.NewResult(1,1))
	if _, err := b.repo.NewRequest(b.id); err != nil {
		t.Error(err.Error())
	}
}

func TestFileConverterData_NewRequest_Fail(t *testing.T) {
	b := BeforeEach(t)
	defer AfterEach(t, b)
	b.mock.ExpectExec(fmt.Sprintf("INSERT INTO %s", tableName)).
		WithArgs(b.id, enums.QUEUED.Name(), "NONE", AnyTime{}).
		WillReturnError(testingError)
	if _, err := b.repo.NewRequest(b.id); err == nil {
		t.Error(errorExpectedError)
	}
}

func TestFileConverterData_StartConversion_Success(t *testing.T) {
	b := BeforeEach(t)
	defer AfterEach(t, b)
	b.mock.ExpectExec(fmt.Sprintf("UPDATE %s", tableName)).
		WithArgs(enums.CONVERTING.Name(), AnyTime{}, b.id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if _, err := b.repo.StartConversion(b.id); err != nil {
		t.Error(err.Error())
	}
}

func TestFileConverterData_StartConversion_Fail(t *testing.T) {
	b := BeforeEach(t)
	defer AfterEach(t, b)
	b.mock.ExpectExec(fmt.Sprintf("UPDATE %s", tableName)).
		WithArgs(enums.CONVERTING.Name(), AnyTime{}, b.id).
		WillReturnError(testingError)
	if _, err := b.repo.StartConversion(b.id); err == nil {
		t.Error(errorExpectedError)
	}
}

func TestFileConverterData_CompleteConversion_Success(t *testing.T) {
	b := BeforeEach(t)
	defer AfterEach(t, b)
	testUrl := "test-url"
	b.mock.ExpectExec(fmt.Sprintf("UPDATE %s", tableName)).
		WithArgs(enums.COMPLETED.Name(), testUrl, AnyTime{}, b.id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if _, err := b.repo.CompleteConversion(b.id, testUrl); err != nil {
		t.Error(err.Error())
	}
}

func TestFileConverterData_CompleteConversion_Fail(t *testing.T) {
	b := BeforeEach(t)
	defer AfterEach(t, b)
	testUrl := "test-url"
	b.mock.ExpectExec(fmt.Sprintf("UPDATE %s", tableName)).
		WithArgs(enums.COMPLETED.Name(), testUrl, AnyTime{}, b.id).
		WillReturnError(testingError)
	if _, err := b.repo.CompleteConversion(b.id, "test-url"); err == nil {
		t.Error(errorExpectedError)
	}
}

func TestFileConverterData_FailConversion_Success(t *testing.T) {
	b := BeforeEach(t)
	defer AfterEach(t, b)
	b.mock.ExpectExec(fmt.Sprintf("UPDATE %s", tableName)).
		WithArgs(enums.FAILED.Name(), AnyTime{}, b.id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if _, err := b.repo.FailConversion(b.id); err != nil {
		t.Error(err.Error())
	}
}

func TestFileConverterData_FailConversion_Fail(t *testing.T) {
	b := BeforeEach(t)
	defer AfterEach(t, b)
	b.mock.ExpectExec(fmt.Sprintf("UPDATE %s", tableName)).
		WithArgs(enums.FAILED.Name(), AnyTime{}, b.id).
		WillReturnError(testingError)
	if _, err := b.repo.FailConversion(b.id); err == nil {
		t.Error(errorExpectedError)
	}
}

func TestFileConverterData_GetConversion_Success(t *testing.T) {
	b := BeforeEach(t)
	defer AfterEach(t, b)
	id := b.id
	status := enums.COMPLETED.Name()
	currUrl := "test-url"
	lastUpdated := time.Now()
	columns := []string{
		"Id",
		"Status",
		"CurrUrl",
		"Last_Updated",
	}
	b.mock.ExpectQuery(fmt.Sprintf("SELECT \\* FROM %s", tableName)).
		WithArgs(b.id).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(id, status, currUrl, lastUpdated))
	res, err := b.repo.GetConversion(id)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, id, res.Id)
	assert.Equal(t, status, res.Status)
	assert.Equal(t, currUrl, res.CurrUrl)
	assert.Equal(t, lastUpdated, res.LastUpdated)
}

func TestFileConverterData_GetConversion_Fail(t *testing.T) {
	b := BeforeEach(t)
	defer AfterEach(t, b)
	b.mock.ExpectQuery(fmt.Sprintf("SELECT \\* FROM %s", tableName)).
		WithArgs(b.id).
		WillReturnError(testingError)
	res, err := b.repo.GetConversion(b.id)
	assert.Nil(t, res)
	assert.NotNil(t, err)
}
