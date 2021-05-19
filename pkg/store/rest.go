package store

import (
	"context"
	"log"
	"net/url"

	"github.com/google/uuid"
	"github.com/simulot/aspiratv/pkg/models"
	"github.com/simulot/aspiratv/pkg/myhttp"
)

const (
	APIURL           = "api/"
	settingsURL      = "settings/"
	subscriptionsURL = "subscriptions/"
)

// API implements a store using RestAPI.
type RestStore struct {
	endPoint string
	client   *myhttp.Client
}

func NewRestStore(endPoint string) *RestStore {
	return &RestStore{
		endPoint: endPoint,
		client: myhttp.NewClient(
			myhttp.WithLogger(log.Default()),
		),
	}
}

func (s *RestStore) GetSettings(ctx context.Context) (models.Settings, error) {
	req, err := s.client.NewRequestJSON(ctx, s.endPoint+settingsURL, nil, nil, nil)
	if err != nil {
		return models.Settings{}, err
	}
	var settings models.Settings
	err = s.client.GetJSON(req, &settings)
	if err != nil {
		return models.Settings{}, err
	}
	return settings, err
}

func (s *RestStore) SetSettings(ctx context.Context, settings models.Settings) (models.Settings, error) {
	req, err := s.client.NewRequestJSON(ctx, s.endPoint+settingsURL, nil, nil, settings)
	if err != nil {
		return models.Settings{}, err
	}
	err = s.client.PostJSON(req, &settings)
	if err != nil {
		return models.Settings{}, err
	}
	return settings, err
}

func (s *RestStore) GetSubscription(ctx context.Context, UUID uuid.UUID) (models.Subscription, error) {
	v := url.Values{}
	v.Set("UUID", UUID.String())
	req, err := s.client.NewRequestJSON(ctx, s.endPoint+subscriptionsURL, nil, v, nil)
	if err != nil {
		return models.Subscription{}, err
	}
	sub := models.Subscription{}
	err = s.client.GetJSON(req, &sub)
	if err != nil {
		return models.Subscription{}, err
	}
	return sub, nil
}
func (s *RestStore) GetAllSubscriptions(ctx context.Context) ([]models.Subscription, error) {
	req, err := s.client.NewRequestJSON(ctx, s.endPoint+subscriptionsURL, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	subs := []models.Subscription{}
	err = s.client.GetJSON(req, &subs)
	return subs, nil
}

func (s *RestStore) SetSubscription(ctx context.Context, sub models.Subscription) (models.Subscription, error) {
	req, err := s.client.NewRequestJSON(ctx, s.endPoint+subscriptionsURL, nil, nil, sub)
	if err != nil {
		return models.Subscription{}, err
	}
	err = s.client.PostJSON(req, &sub)

	return sub, nil
}
