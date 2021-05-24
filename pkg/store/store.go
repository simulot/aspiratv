package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/simulot/aspiratv/pkg/models"
)

var ErrorNotFound = errors.New("ressource not found")

type Store interface {
	SettingsInterface
	SubscriptionInterfaces
}

type SettingsInterface interface {
	GetSettings(context.Context) (models.Settings, error)
	SetSettings(context.Context, models.Settings) (models.Settings, error)
}

type SubscriptionInterfaces interface {
	GetSubscription(context.Context, uuid.UUID) (models.Subscription, error)
	DeleteSubscription(context.Context, uuid.UUID) error
	GetAllSubscriptions(context.Context) ([]models.Subscription, error)
	SetSubscription(context.Context, models.Subscription) (models.Subscription, error)
}
