package frontend

import (
	"log"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/simulot/aspiratv/pkg/frontend/bulma"
	"github.com/simulot/aspiratv/pkg/models"
)

type SubscriptionPage struct {
	app.Compo
	Subs    map[string]models.Subscription
	Editing string
}

func newSubscriptionPage() app.Composer {
	return &SearchPage{}
}

func (s *SubscriptionPage) OnMount(ctx app.Context) {
	subs, err := MyAppState.API.GetAllSubscriptions(ctx)
	if err != nil {
		log.Printf("SubscriptionPage->GetAllSubscriptions: %s", err)
		// MyAppState.Dispatch.Publish(models.NewMessage(fmt.Sprintf("Erreur :%s", err)).SetPinned(true).SetStatus(models.StatusInfo))
		return
	}
	s.Subs = map[string]models.Subscription{}
	for _, sub := range subs {
		s.Subs[sub.UUID.String()] = sub
	}
}

func (s *SubscriptionPage) Render() app.UI {
	MyAppState.CurrentPage = PageSubscriptions
	if s.Subs == nil {
		return app.Text("Initialisation")
	}

	return app.Div().Body(
		app.H1().Class("title is-1").Text("Abonnements"),
		app.If(s.Editing != "", s.RenderSubscription()).
			Else(app.Range(s.Subs).Map(func(k string) app.UI {
				return app.P().Text(s.Subs[k].Title)
			})),
	)
}

func (s *SubscriptionPage) RenderSubscription() app.UI {
	sub := s.Subs[s.Editing]
	active := map[bool]string{false: "disabled", true: "enabled"}[sub.Enabled]

	return app.Div().Body(
		app.Button().Class("button").Text("Retour"),
		bulma.NewTextField(sub.Title, "Nom de l'abonnement", "Nom").WithOnInput(func(v string) {
			sub.Title = v
		}),
		bulma.NewRadioFields(active, "Abonnement activé").
			WithOption("disabled", "Desactivé", !sub.Enabled).
			WithOption("enabled", "Activé", sub.Enabled).
			WhitOnInput(func(v string) {
				sub.Enabled = v == "enabled"
			}),
		bulma.NewTextField(sub.Title, "Page de l'émission", "URL").WithOnInput(func(v string) {
			sub.ShowPageURL = v
		}),
	)

}
