package models

import (
	"github.com/google/uuid"
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

type TypeDownloader int

const (
	DownloaderFFMPEG TypeDownloader = iota
)

type DownloadItem struct {
	Downloader TypeDownloader // Providers decides the downloader
	MediaInfo  MediaInfo      // Details of media, contains the stream's url
}
