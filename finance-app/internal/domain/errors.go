package domain

import "errors"

// Common sentinel errors used across the application.
var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("already exists")
)
