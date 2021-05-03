package frontend

import (
	"path/filepath"

	"github.com/google/uuid"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/simulot/aspiratv/pkg/models"
)

type DownloadDialog struct {
	app.Compo

	Result models.SearchResult
	Path   string
	Back   func()
}

func NewDownloadDialog(r models.SearchResult, back func()) *DownloadDialog {
	return &DownloadDialog{
		Result: r,
		Back:   back,
		Path:   filepath.Join(MyAppState.Settings.LibraryPath, r.Show),
	}
}

func (dd *DownloadDialog) Render() app.UI {
	return app.Div().Class("modal is-active").
		Body(
			app.Div().Class("modal-background"),
			app.Div().Class("modal-card").Body(
				app.Header().Class("modal-card-head").Body(
					app.P().Class("modal-card-title").Text("Téléchargement"),
					app.Button().Class("delete").Aria("label", "close").OnClick(dd.OnBack),
				),
				app.Section().Class("modal-card-body").Body(
					app.Div().Class("field").Body(
						app.Label().Class("label").Text("Répertoire sur le serveur"),
						app.Div().Class("control").Body(
							app.Input().Type("text").Placeholder("Chemin").OnChange(dd.ValueTo(&dd.Path)).Value(dd.Path),
						),
					),
				),
				app.Footer().Class("modal-card-foot").Body(
					app.Button().Class("button is-success").Text("Télécharger").OnClick(dd.OnDownload),
					app.Button().Text("Annuler").OnClick(dd.OnBack),
				),
			),
		)
}

func (dd *DownloadDialog) OnBack(ctx app.Context, e app.Event) {
	dd.Back()
}
func (dd *DownloadDialog) OnDownload(ctx app.Context, e app.Event) {
	var err error
	task := models.DownloadTask{
		ID:     uuid.New(),
		Path:   dd.Path,
		Result: dd.Result,
	}

	task, err = MyAppState.API.PostDownload(ctx, task)
	if err != nil {
		MyAppState.Dispatch.Publish(models.NewNotification(err.Error(), models.NotificationError))
		return
	}
	MyAppState.Dispatch.Publish(models.NewNotification("Démarrage du téléchargement", models.NotificationInfo))
}
