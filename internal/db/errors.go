package db

import "errors"

// ErrNotFound is returned by Store methods when a queried row does not exist.
var ErrNotFound = errors.New("not found")
