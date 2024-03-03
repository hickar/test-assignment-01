package domain

import "errors"

var ErrNotFound = errors.New("queried entity not found")

var ErrAlreadyProcessed = errors.New("incoming event was already processed")

var ErrInvalidData = errors.New("invalid input data")
