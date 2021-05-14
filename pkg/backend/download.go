package backend

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/simulot/aspiratv/pkg/library"
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

	settings, err := s.store.GetSettings()
	if p == nil {
		s.sendError(w, APIError{err: err, code: http.StatusInternalServerError})
		return
	}

	var fileNamer models.FileNamer
	switch task.Result.Type {
	case models.TypeCollection:
		fileNamer = settings.DefaultCollectionSettings.FileNamer
	case models.TypeSeries:
		fileNamer = settings.DefaultSeriesSettings.FileNamer
	case models.TypeTVShow:
		fileNamer = settings.DefaultTVShowsSettings.FileNamer
	}

	go library.NewBatchDownloader(
		task.Result.Title,
		settings.LibraryPath,
		fileNamer,
	).WithLogger(log.Default()).WithPublisher(s.dispatcher).Download(s.backgroundCtx, c)
	s.writeJsonResponse(w, task, http.StatusOK)

}
