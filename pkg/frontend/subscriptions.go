package frontend

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/simulot/aspiratv/pkg/models"
)

type SubscriptionPage struct {
	app.Compo
	Subs map[uuid.UUID]models.Subscription
}

func (s *SubscriptionPage) OnMount(ctx app.Context) {
	subs, err := MyAppState.API.GetAllSubscriptions(ctx)
	if err != nil {
		MyAppState.Dispatch.Publish(models.NewMessage(fmt.Sprintf("Erreur :%s", err)).SetPinned(true).SetStatus(models.StatusInfo))
		return
	}
	s.Subs = map[uuid.UUID]models.Subscription{}
	for _, sub := range subs {
		s.Subs[sub.UUID] = sub
	}
}

func (s *SubscriptionPage) Render() app.UI {
	MyAppState.CurrentPage = PageSubscriptions
	if s.Subs == nil {
		return app.Text("Initialisation")
	}

	return AppPageRender((app.H1().Class("title is-1").Text("Abonnements")))

}
