package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/simulot/aspiratv/pkg/dispatcher"
	"github.com/simulot/aspiratv/pkg/download"
	"github.com/simulot/aspiratv/pkg/models"
)

/*
	Download Handler
	GET  --> List of ongoing download
	POST --> Add a Download
	DELETE --> Kill a download
*/

func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	// case http.MethodGet:
	// 	s.getDownloads(w, r)
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		s.postDownload(w, r)
	default:
		s.sendError(w, APIError{nil, http.StatusMethodNotAllowed, ""})
	}
}

// func (s *Server) getDownloads(w http.ResponseWriter, r *http.Request) {
// s.sendError(w, APIError{code: http.StatusNotImplemented})
// }

func (s *Server) postDownload(w http.ResponseWriter, r *http.Request) {
	var task models.DownloadTask
	var err error

	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		s.sendError(w, err)
		return
	}
	// log.Printf("[HTTPSERVER] DownloadTask: %#v", task)

	p := s.Provider(task.Result.Provider)
	if p == nil {
		s.sendError(w, APIError{code: http.StatusBadRequest, message: "Unknown provider"})
		return
	}

	c, err := p.GetMedias(s.backgroundCtx, task)

	if p == nil {
		s.sendError(w, APIError{err: err, code: http.StatusBadRequest})
		return
	}

	go s.GetMedias(task, c)
	s.writeJsonResponse(w, task, http.StatusOK)

}

func (s *Server) GetMedias(task models.DownloadTask, c <-chan models.DownloadItem) {
	// jobChannel := make(chan job.Task, 1)
	// job := job.NewJob()
	// go job.Run(s.backgroundCtx, jobChannel)
	// defer func() {
	// 	close(jobChannel)
	// 	job.End()
	// }()

	jobDownload := models.NewMessage(fmt.Sprintf("Téléchargement de %q", task.Result.Show)).SetPinned(true).SetStatus(models.StatusInfo)
	s.dispatcher.Publish(jobDownload)

	dispatchError := func(err error) {
		jobDownload.SetText(fmt.Sprintf("Téléchargement de %q: erreur: %s", task.Result.Show, err)).SetStatus(models.StatusError).SetPinned(true)
		s.dispatcher.Publish(jobDownload)
	}

	settings, err := s.store.GetSettings()
	if err != nil {
		dispatchError(err)
		return
	}

	for {
		log.Printf("GetMedias Wait for a new media")
		select {
		case item, ok := <-c:
			if !ok {
				jobDownload.SetText(fmt.Sprintf("Téléchargement de %q terminé", task.Result.Show)).SetStatus(models.StatusSuccess).SetPinned(false)
				s.dispatcher.Publish(jobDownload)
				return
			}
			// itemJob := func() error {
			showPath := path.Join(download.PathClean(settings.LibraryPath), task.Result.Show)
			err = os.MkdirAll(showPath, 0777)
			if err != nil {
				dispatchError(err)
				continue
				// return err
			}
			seasonPath := path.Join(showPath, fmt.Sprintf("Season %02d", item.MediaInfo.Season))
			err = os.MkdirAll(seasonPath, 0777)
			if err != nil {
				dispatchError(err)
				continue
				// return err
			}
			mp4Name := fmt.Sprintf("%s S%02dE%02d %s.mp4", item.MediaInfo.Show, item.MediaInfo.Season, item.MediaInfo.Episode, item.MediaInfo.Title)
			episodePath := path.Join(seasonPath, mp4Name)
			itemProgression := dlProgression{
				Message: models.NewProgression(mp4Name, 0, 0).SetPinned(true).SetStatus(models.StatusInfo),
				d:       s.dispatcher,
			}

			s.dispatcher.Publish(itemProgression.Message)
			item.Downloader.WithProgresser(&itemProgression)
			err := item.Downloader.Download(s.backgroundCtx, episodePath)
			// return err
			if err != nil {
				dispatchError(err)
			}

			// }
			// jobChannel <- itemJob
			log.Printf("au suivant")
		case <-s.backgroundCtx.Done():
			dispatchError(s.backgroundCtx.Err())
			return
		}
	}
}

type dlProgression struct {
	*models.Message
	d *dispatcher.Dispatcher
}

func (p *dlProgression) Progress(current int, total int) {
	p.Message.Progression.Progress(current, total)
	if current >= total {
		p.Status = models.StatusSuccess
		p.SetPinned(false)
	}
	p.d.Publish(p.Message)
}
