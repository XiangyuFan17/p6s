package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"p6s/internal/model"
)

// PostgresDB wraps PostgreSQL database connection and operations
type PostgresDB struct {
	db *sql.DB
}

// NewPostgresDB creates a new PostgresDB instance
func NewPostgresDB() *PostgresDB {
	return &PostgresDB{}
}

// Connect establishes connection to PostgreSQL database
func (p *PostgresDB) Connect(connStr string) error {

	connStr += "&connect_timeout=3"
	
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}


	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetConnMaxIdleTime(time.Second * 30)


	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("database connection test failed: %v", err)
	}

	p.db = db
	return nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// IsConnected checks if database is connected
func (p *PostgresDB) IsConnected() bool {
	return p.db != nil
}

// GetConnections retrieves current database connection information
func (p *PostgresDB) GetConnections(filterType string) ([]model.Connection, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	var query string
	switch filterType {
	case "all":
		query = `SELECT pid, usename, datname, client_addr, application_name, backend_start, state, query 
				FROM pg_stat_activity 
				WHERE pid <> pg_backend_pid() 
				ORDER BY backend_start DESC`
	case "active":
		query = `SELECT pid, usename, datname, client_addr, application_name, backend_start, state, query 
				FROM pg_stat_activity 
				WHERE pid <> pg_backend_pid() AND state = 'active' 
				ORDER BY backend_start DESC`
	case "blocked":
		query = `SELECT blocked_activity.pid, 
					blocked_activity.usename, 
					blocked_activity.datname, 
					blocked_activity.client_addr, 
					blocked_activity.application_name, 
					blocked_activity.backend_start, 
					blocked_activity.state, 
					blocked_activity.query 
				FROM pg_stat_activity blocked_activity 
				JOIN pg_locks blocked_locks ON blocked_activity.pid = blocked_locks.pid 
				JOIN pg_locks blocking_locks ON blocked_locks.transactionid = blocking_locks.transactionid AND blocked_locks.pid != blocking_locks.pid 
				JOIN pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid 
				WHERE NOT blocked_locks.granted 
				ORDER BY blocked_activity.backend_start DESC`
	default:
		return nil, fmt.Errorf("unknown filter type: %s", filterType)
	}


	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()


	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query connection info: %v", err)
	}
	defer rows.Close()

	var connections []model.Connection
	for rows.Next() {
		var conn model.Connection
		if err := rows.Scan(
			&conn.PID,
			&conn.User,
			&conn.Database,
			&conn.ClientAddr,
			&conn.ApplicationName,
			&conn.BackendStart,
			&conn.State,
			&conn.Query,
		); err != nil {
			return nil, fmt.Errorf("failed to parse connection info: %v", err)
		}
		connections = append(connections, conn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate connection info: %v", err)
	}

	return connections, nil
}

// GetTableStats retrieves table size statistics
func (p *PostgresDB) GetTableStats() ([]model.TableStat, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	query := `SELECT 
			n.nspname as schema,
			c.relname as name,
			pg_size_pretty(pg_total_relation_size(c.oid)) as total_size,
			pg_size_pretty(pg_relation_size(c.oid)) as table_size,
			pg_size_pretty(pg_indexes_size(c.oid)) as index_size,
			c.reltuples::bigint as row_count
		FROM pg_class c
		LEFT JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relkind = 'r'
		AND n.nspname NOT IN ('pg_catalog', 'information_schema')
		ORDER BY pg_total_relation_size(c.oid) DESC
		LIMIT 100`


	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()


	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table statistics: %v", err)
	}
	defer rows.Close()

	var tableStats []model.TableStat
	for rows.Next() {
		var stat model.TableStat
		if err := rows.Scan(
			&stat.Schema,
			&stat.Name,
			&stat.TotalSize,
			&stat.TableSize,
			&stat.IndexSize,
			&stat.RowCount,
		); err != nil {
			return nil, fmt.Errorf("failed to parse table statistics: %v", err)
		}
		tableStats = append(tableStats, stat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate table statistics: %v", err)
	}

	return tableStats, nil
}

// GetDatabaseVersion retrieves database version information
func (p *PostgresDB) GetDatabaseVersion() (string, error) {
	if p.db == nil {
		return "", fmt.Errorf("database not connected")
	}

	var version string
	if err := p.db.QueryRow("SELECT version()").Scan(&version); err != nil {
		return "", fmt.Errorf("failed to get database version: %v", err)
	}

	return version, nil
}

// GetCurrentDatabase retrieves current database name
func (p *PostgresDB) GetCurrentDatabase() (string, error) {
	if p.db == nil {
		return "", fmt.Errorf("database not connected")
	}


	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dbName string
	if err := p.db.QueryRowContext(ctx, "SELECT current_database()").Scan(&dbName); err != nil {
		return "", fmt.Errorf("failed to get current database name: %v", err)
	}

	return dbName, nil
}

// GetDatabases retrieves list of all databases
func (p *PostgresDB) GetDatabases() ([]string, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	query := `SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname`

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query database list: %v", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, fmt.Errorf("failed to parse database name: %v", err)
		}
		databases = append(databases, dbName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate database list: %v", err)
	}

	return databases, nil
}

// ExecuteCustomQuery executes custom SQL query
func (p *PostgresDB) ExecuteCustomQuery(sqlQuery string) ([][]interface{}, []string, error) {
	if p.db == nil {
		return nil, nil, fmt.Errorf("database not connected")
	}


	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()


	rows, err := p.db.QueryContext(ctx, sqlQuery)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute SQL query: %v", err)
	}
	defer rows.Close()


	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get column info: %v", err)
	}


	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get column types: %v", err)
	}


	var results [][]interface{}


	for rows.Next() {

		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}


		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, nil, fmt.Errorf("failed to scan row data: %v", err)
		}


		row := make([]interface{}, len(columns))
		for i, val := range values {
			if val == nil {
				row[i] = "NULL"
			} else {
		
				switch columnTypes[i].DatabaseTypeName() {
				case "VARCHAR", "TEXT", "CHAR":
					if b, ok := val.([]byte); ok {
						row[i] = string(b)
					} else {
						row[i] = fmt.Sprintf("%v", val)
					}
				case "INT4", "INT8", "BIGINT", "INTEGER":
					row[i] = fmt.Sprintf("%v", val)
				case "FLOAT4", "FLOAT8", "NUMERIC", "DECIMAL":
					row[i] = fmt.Sprintf("%v", val)
				case "BOOL":
					row[i] = fmt.Sprintf("%v", val)
				case "TIMESTAMP", "TIMESTAMPTZ", "DATE", "TIME":
					if t, ok := val.(time.Time); ok {
						row[i] = t.Format("2006-01-02 15:04:05")
					} else {
						row[i] = fmt.Sprintf("%v", val)
					}
				default:
					if b, ok := val.([]byte); ok {
						row[i] = string(b)
					} else {
						row[i] = fmt.Sprintf("%v", val)
					}
				}
			}
		}

		results = append(results, row)
	}


	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to iterate results: %v", err)
	}

	return results, columns, nil
}