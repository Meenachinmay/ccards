package transaction

import (
	"database/sql"
)

type repository struct {
	db *sql.DB
}

// NewRepository creates a new transaction repository
func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

// Implement the Repository interface methods here