package model

import (
	"database/sql"
	"time"
)

// Connection represents database connection information
type Connection struct {
	PID            int
	User           sql.NullString
	Database       sql.NullString
	ClientAddr     sql.NullString
	ApplicationName sql.NullString
	BackendStart   time.Time
	State          sql.NullString
	Query          sql.NullString
}

// TableStat represents table statistics information
type TableStat struct {
	Schema     string
	Name       string
	TotalSize  string
	TableSize  string
	IndexSize  string
	RowCount   int64
}