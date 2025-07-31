package auth

import (
	"database/sql"
)

type repository struct {
	db *sql.DB
}

// NewRepository creates a new auth repository
func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}

// Implement the Repository interface methods here