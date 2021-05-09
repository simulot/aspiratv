package models

import (
	"github.com/google/uuid"
	"github.com/simulot/aspiratv/pkg/download"
)

type DownloadTask struct {
	ID     uuid.UUID
	DryRun bool         // True for testing parameters
	Path   string       // Root path for the download
	Result SearchResult // Show / Series to be downloaded
	Status struct {
		Files        int // Files downloaded
		Bytes        int // Bytes downloaded
		CurrentSpeed int // Bytes per second
	}
}

type DownloadItem struct {
	Downloader download.Downloader
	MediaInfo  MediaInfo // Details of media, contains the stream's url
}
