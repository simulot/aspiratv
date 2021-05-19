package store

import (
	"errors"

	"github.com/google/uuid"
	"github.com/simulot/aspiratv/pkg/models"
)

var ErrorNotFound = errors.New("ressource not found")

type Store interface {
	ProviderInterface
}

type ProviderInterface interface {
	GetSettings() (models.Settings, error)
	SetSettings(models.Settings) (models.Settings, error)
}

type SubscriptionInterfaces interface {
	GetSubscription(id uuid.UUID) (models.Subscription, error)
	GetAllSubscription() ([]models.Subscription, error)
	SetSubscription(models.Subscription) (models.Subscription, error)
}
