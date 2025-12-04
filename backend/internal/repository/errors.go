// Package repository contains shared errors for data access layer
package repository

import "errors"

var (
	// ErrNotFound indicates that a record was not found in the database
	ErrNotFound = errors.New("record not found")
	// ErrDuplicate indicates a unique constraint violation
	ErrDuplicate = errors.New("duplicate record")
)
