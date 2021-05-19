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
	path := MyAppState.Settings.LibraryPath
	namer := models.DefaultFileNamer[models.PathNamingType(r.Type)]
	if namer != nil {
		path = filepath.Join(path, namer.ShowPathString(models.MediaInfo{
			Type:     r.Type,
			Title:    r.Title,
			Show:     r.Show,
			Aired:    r.Aired,
			Year:     r.Aired.Year(),
			Channel:  r.Chanel,
			Provider: r.Provider,
		}))
	}

	return &DownloadDialog{
		Result: r,
		Back:   back,
		Path:   path,
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
							app.Input().Type("text").Placeholder("Chemin").OnChange(dd.ValueTo(&dd.Path)).Value(dd.Path).Disabled(true),
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
		MyAppState.Dispatch.Publish(models.NewMessage(err.Error()).SetStatus(models.StatusError))
		return
	}
	dd.Back()
}
