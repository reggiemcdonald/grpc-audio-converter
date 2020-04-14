// Simple crud operations for managing new and existing file conversion requests
package converterservice

import (
	"database/sql"
	"fmt"
	"github.com/reggiemcdonald/grpc-audio-converter/pb"
	"log"
	"time"
)

type FileConverterData struct {
	db *sql.DB
}

/*
 * Database constants
 */
const (
	tableName  = "convert_jobs"
)

/*
 * FileConverterData constructor
 */
func NewFileConverterData(dbUser string, dbPass string) *FileConverterData {
	connstr := fmt.Sprintf("user=%s password=%s sslmode=disable", dbUser, dbPass)
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
 *   id string PRIMARY_KEY
 *   status string [QUEUED | CONVERTING | COMPLETED | FAILED]
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
 * Updates the status of the current file conversion to complete, including the presigned URL to the bucket object
 */
func (f *FileConverterData) CompleteConversion(id string, url string) (bool, error) {
	stmt := fmt.Sprintf("UPDATE %s SET status=$1, curr_url=$2, last_updated=$3 WHERE id=$4", tableName)
	status, lastUpdated := pb.ConvertFileQueryResponse_COMPLETED.String(), time.Now()
	_, err := f.db.Exec(stmt, status, url, lastUpdated, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

/*
 * Updates the status of the specified file conversion to failed, along with the timestamp of failure
 */
func (f *FileConverterData) FailConversion(id string) (bool, error) {
	stmt := fmt.Sprintf("UPDATE %s SET status=$1, last_updated=$2 WHERE id=%3", tableName)
	status, lastUpdated := pb.ConvertFileQueryResponse_FAILED.String(), time.Now()
	_, err := f.db.Exec(stmt, status, lastUpdated, id)
	if err != nil {
		return false, nil
	}
	return true, nil
}