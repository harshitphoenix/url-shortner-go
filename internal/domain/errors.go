package domain

import "errors"

var (
	ErrNotFound     = errors.New("url not found")
	ErrInvalidURL   = errors.New("invalid url")
	ErrCodeConflict = errors.New("code already exists")
	ErrInvalidCode  = errors.New("invalid code")
)
