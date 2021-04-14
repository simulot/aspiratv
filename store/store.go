package store

import (
	"errors"
)

type Store interface {
}

var ErrorNotFound = errors.New("ressource not found")
