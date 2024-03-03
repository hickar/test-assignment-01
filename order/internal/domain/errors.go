package domain

import "errors"

var ErrNotFound = errors.New("queried order is not found")

var ErrInvalidData = errors.New("invalid input data provided")
