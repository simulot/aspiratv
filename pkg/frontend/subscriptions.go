package frontend

import (
	"fmt"
	"log"
	"strconv"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/simulot/aspiratv/pkg/frontend/bulma"
	"github.com/simulot/aspiratv/pkg/models"
)

type SubscriptionPage struct {
	app.Compo
	SubPage subscriptionSubPage
	Subs    map[string]models.Subscription
	Sub     models.Subscription
}

type subscriptionSubPage int

const (
	subscriptionPageList subscriptionSubPage = iota
	subscriptionPageEdit
)

func newSubscriptionPage(action app.Action) app.Composer {
	s := SubscriptionPage{}

	if action.Value != nil {
		if result, ok := action.Value.(models.SearchResult); ok {
			s.SubPage = subscriptionPageEdit
			s.Sub = models.Subscription{
				Title:       result.Show,
				ShowID:      result.ID,
				PollRhythm:  models.RhythmWeekly,
				ShowType:    result.Type,
				Provider:    result.Provider,
				LimitNumber: 20,
				Enabled:     true,
				ShowPageURL: result.PageURL,
			}
		}
	}

	return &s
}

func (s *SubscriptionPage) loadSubs(ctx app.Context) {
	subs, err := MyAppState.API.GetAllSubscriptions(ctx)
	if err != nil {
		log.Printf("SubscriptionPage->GetAllSubscriptions: %s", err)
		// MyAppState.Dispatch.Publish(models.NewMessage(fmt.Sprintf("Erreur :%s", err)).SetPinned(true).SetStatus(models.StatusInfo))
		return
	}
	list := map[string]models.Subscription{}
	for _, sub := range subs {
		list[sub.UUID.String()] = sub
	}
	s.Subs = list
}

func (s *SubscriptionPage) OnMount(ctx app.Context) {
	if s.Subs == nil {
		s.loadSubs(ctx)
		s.Update()
	}
}

func (s *SubscriptionPage) Render() app.UI {
	log.Printf("SubscriptionPage.Render s:%p s.Suv:%#v", s, s.Sub)
	return app.Div().Body(
		app.H1().Class("title is-1").Text("Abonnements"),
		app.If(s.SubPage != subscriptionPageList,
			newEditSubscriptionPage(&s.Sub)).
			Else(NewSubscriptionList(s.Subs)),
	)
}

type SubscriptionList struct {
	app.Compo
	Subs map[string]models.Subscription
}

func NewSubscriptionList(subs map[string]models.Subscription) *SubscriptionList {
	return &SubscriptionList{
		Subs: subs,
	}
}
func (l *SubscriptionList) Render() app.UI {
	if l.Subs == nil {
		return app.Text("Loading")
	}
	return app.Div().Class("columns").Body(
		app.Range(l.Subs).Map(func(k string) app.UI {
			s := l.Subs[k]
			log.Printf("k:%s ,v:%s", k, s.Title)
			return app.Div().Class("column").Body(
				app.H6().Class("title is-6").Text(s.Title),
				app.P().Text(l.Subs[k].Title),
				app.P().Body(
					app.If(s.Enabled, app.Text("Activé")).Else(app.Text("Désactivé")),
				),
				app.P().Body(
					app.Text("Fournisseur de contenu : "),
					app.Text(s.Provider),
				),
				app.P().Body(
					app.Text("Dernière interrogation le : "),
					app.If(s.LastRun.IsZero(), app.Text("jamais interrogé")).Else(app.Text(s.LastRun.Format("02/01/2006 à 15:04"))),
					app.P().Body(
						app.Text("Dernière vidéo collectée le : "),
						app.If(s.LastSeenMedia.IsZero(), app.Text("aucune video")).Else(app.Text(s.LastSeenMedia.Format("02/01/2006 à 15:04"))),
					),
				),
				app.Div().Class("field is-grouped").Body(
					app.Div().Class("control").Body(
						app.Button().Class("button is-link").Text("Interroger maintenant"),
					),
					app.Div().Class("control").Body(
						app.Button().Class("button is-link is-light").Text("Modifier"),
					),
				),
			)
		}),
	)
}

type EditSubscriptionPage struct {
	app.Compo
	sub          *models.Subscription
	deleteOpened bool
}

func newEditSubscriptionPage(sub *models.Subscription) *EditSubscriptionPage {
	return &EditSubscriptionPage{
		sub: sub,
	}
}

func (s *EditSubscriptionPage) Render() app.UI {
	active := map[bool]string{false: "disabled", true: "enabled"}[s.sub.Enabled]
	log.Printf("UUID :%s", s.sub.UUID)

	return app.Div().Body(
		app.Button().Class("button").Text("Retour").OnClick(func(ctx app.Context, e app.Event) {
			GotoPage(ctx, PageSearchOnLine, 0, nil)
		}),
		bulma.NewTextField(s.sub.Title, "Nom de l'abonnement", "Nom").WithOnInput(func(v string) {
			s.sub.Title = v
		}),
		app.Div().Body(
			app.Label().Class("label").Text("Informations :"),
			app.P().Body(
				app.Text("Fournisseur de contenu : "),
				app.Text(s.sub.Provider),
			),
			app.P().Body(
				app.Text("Dernière interrogation le : "),
				app.If(s.sub.LastRun.IsZero(), app.Text("jamais interrogé")).Else(app.Text(s.sub.LastRun.Format("02/01/2006 à 15:04"))),
				app.P().Body(
					app.Text("Dernière vidéo collectée le : "),
					app.If(s.sub.LastSeenMedia.IsZero(), app.Text("aucune video")).Else(app.Text(s.sub.LastSeenMedia.Format("02/01/2006 à 15:04"))),
				),
			),
			bulma.NewRadioFields(active, "Activation").
				WithOption("disabled", "Désactivé", !s.sub.Enabled).
				WithOption("enabled", "Activé", s.sub.Enabled).
				WhitOnInput(func(v string) {
					s.sub.Enabled = v == "enabled"
				}),
			bulma.NewTextField(s.sub.ShowPageURL, "Page de l'émission", "URL").WithOnInput(func(v string) {
				s.sub.ShowPageURL = v
			}),

			bulma.NewSelectField(s.sub.PollRhythm.String(), "Interroger le serveur :").
				WithOption(models.RhythmDaily.String(), "Tous les jours", s.sub.PollRhythm == models.RhythmDaily).
				WithOption(models.RhythmWeekly.String(), "Toutes les semaines", s.sub.PollRhythm == models.RhythmWeekly).
				WithOption(models.RhythmMonthly.String(), "Tous les mois", s.sub.PollRhythm == models.RhythmMonthly).
				WhitOnInput(func(v string) {
					s.sub.PollRhythm, _ = models.PoolRhythmTypeString(v)
				}),

			bulma.NewTextField(itos(s.sub.LimitNumber), "Nombre maximal de vidéos à collecter à chaque essai :", "Nombre de videos").
				WithOnInput(func(v string) {
					s.sub.LimitNumber = stoi(v)
				}),

			bulma.NewTextField(itos(s.sub.MaxAge), "Exclure les vidéos diffusées depuis plus de X jours:", "Nombre de jours").
				WithOnInput(func(v string) {
					s.sub.MaxAge = stoi(v)
				}),

			bulma.NewTextField(itos(s.sub.DeleteAge), "Supprimer les vidéos téléchagées au delà de X jours:", "Nombre de jours").
				WithOnInput(func(v string) {
					s.sub.DeleteAge = stoi(v)
				}),
		),

		app.Div().Class("field is-grouped").Body(
			app.Div().Class("control").Body(
				app.Button().Class("button is-link").Text("Enregistrer").OnClick(s.submit),
			),
			app.Div().Class("control").Body(
				app.Button().Class("button is-link is-light").Text("Annuler").OnClick(s.cancel),
			),
			app.Div().Class("control").Body(
				app.Button().Class("button is-link is-danger").Text("Supprimer").OnClick(s.delete),
			),
		),
		app.If(s.deleteOpened,
			bulma.NewConfirmDialog("Supprimer l'abonnement", "Supprimer", func(ctx app.Context, action bool) {
				s.deleteOpened = false
				if action {
					err := MyAppState.API.Store.DeleteSubscription(ctx, s.sub.UUID)
					if err != nil {
						MyAppState.Dispatch.Publish(models.NewMessage(fmt.Sprintf("Erreur : %s", err.Error())).SetPinned(true).SetStatus(models.StatusError))
						return
					}
					// TODO GoBack
				}
			}),
		),
	)
}

func (s *EditSubscriptionPage) submit(ctx app.Context, e app.Event) {
	resp, err := MyAppState.API.SetSubscription(ctx, *s.sub)
	if err != nil {
		MyAppState.Dispatch.Publish(models.NewMessage(fmt.Sprintf("Erreur : %s", err.Error())).SetPinned(true).SetStatus(models.StatusError))
	}
	*s.sub = resp
}
func (s *EditSubscriptionPage) cancel(ctx app.Context, e app.Event) {
	// TODO: goback
}

func (s *EditSubscriptionPage) delete(ctx app.Context, e app.Event) {
	s.deleteOpened = true
}

func itos(i int) string {
	return strconv.Itoa(i)
}

func stoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
