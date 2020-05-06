// Simple crud operations for managing new and existing file conversion requests
package converterservice

import (
	"database/sql"
	"fmt"
	"github.com/reggiemcdonald/grpc-audio-converter/pb"
	"log"
	"time"
)

type FileConverterDataRepository interface {
	NewRequest(id string) (bool, error)
	CompleteConversion(id string, url string) (bool, error)
	FailConversion(id string) (bool, error)
	GetConversion(id string) (*ConvertJob, error)
}

type FileConverterData struct {
	db *sql.DB
}

type ConvertJob struct {
	Id          string
	Status      string
	CurrUrl     string
	LastUpdated time.Time
}

/*
 * Database constants
 */
const (
	host       = "converter_db"
	tableName  = "convert_jobs"
)

/*
 * FileConverterData constructor
 */
func NewFileConverterData(dbUser string, dbPass string) *FileConverterData {
	connstr := fmt.Sprintf("host=%s user=%s password=%s sslmode=disable", host, dbUser, dbPass)
	log.Print(connstr)
	db, err := sql.Open("postgres", connstr)
	if err != nil {
		log.Fatalf("failed to connect to database encountered, %v", err)
	}
	return &FileConverterData{
		db: db,
	}
}

/*
 * Inserts a request into the database
 * SCHEMA:
 *   Id string PRIMARY_KEY
 *   Status string [QUEUED | CONVERTING | COMPLETED | FAILED]
 *   curr_url string
 *   last_updated timestamp
 */
func (f *FileConverterData) NewRequest(id string) (bool, error) {
	stmt := fmt.Sprintf("INSERT INTO %s VALUES ($1, $2, $3, $4)", tableName)
	status, url, lastTime := pb.ConvertFileQueryResponse_CONVERTING.String(), "NONE", time.Now()
	_, err := f.db.Exec(stmt, id, status, url, lastTime)
	if err != nil {
		return false, err
	}
	return true, nil
}

/*
 * Updates the Status of the current file conversion to complete, including the presigned URL to the bucket object
 */
func (f *FileConverterData) CompleteConversion(id string, url string) (bool, error) {
	stmt := fmt.Sprintf("UPDATE %s SET Status=$1, curr_url=$2, last_updated=$3 WHERE Id=$4", tableName)
	status, lastUpdated := pb.ConvertFileQueryResponse_COMPLETED.String(), time.Now()
	_, err := f.db.Exec(stmt, status, url, lastUpdated, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

/*
 * Updates the Status of the specified file conversion to failed, along with the timestamp of failure
 */
func (f *FileConverterData) FailConversion(id string) (bool, error) {
	stmt := fmt.Sprintf("UPDATE %s SET Status=$1, last_updated=$2 WHERE Id=$3", tableName)
	status, lastUpdated := pb.ConvertFileQueryResponse_FAILED.String(), time.Now()
	_, err := f.db.Exec(stmt, status, lastUpdated, id)
	if err != nil {
		return false, nil
	}
	return true, nil
}

/*
 * Fetches ConvertJob from the database
 */
func (f *FileConverterData) GetConversion(id string) (*ConvertJob, error) {
	stmt := fmt.Sprintf("SELECT * FROM %s WHERE Id=$1", tableName)
	var (
		status string
		currUrl string
		lastUpdated time.Time
	)
	err := f.db.QueryRow(stmt, id).Scan(&id, &status, &currUrl, &lastUpdated)
	if err != nil {
		return nil, err
	}
	return &ConvertJob{id, status, currUrl, lastUpdated}, nil
}