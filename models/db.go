package models

import "database/sql"

// AppDB holds the db connection pool
// All DB access will be implemented as methods on this struct
type AppDB struct {
	DB *sql.DB
}
