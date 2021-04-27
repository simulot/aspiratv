package store

import (
	"errors"

	"github.com/simulot/aspiratv/models"
)

var ErrorNotFound = errors.New("ressource not found")

type Store interface {
	ProviderSettings
}

type ProviderSettings interface {
	GetSettings() (models.Settings, error)
	SetSettings(models.Settings) (models.Settings, error)
}
