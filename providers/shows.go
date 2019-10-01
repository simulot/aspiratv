package providers

import (
	"time"
)

// Show structure handle show details.
// It is shared among all packages
type Show struct {
	ID            string        // Show ID
	Title         string        // Episode title
	HasEpisodes   bool          // This show is a part of a serie of shows
	Show          string        // Show name
	Pitch         string        // Pitch on the episode
	SeasonPitch   string        // Pitch for the season
	Season        string        // Season
	Episode       string        // Episode
	Channel       string        // Channel
	AirDate       time.Time     // Broadcasting date
	Duration      time.Duration // Duration of the show
	StreamURL     string        // url of the video stream
	ThumbnailURL  string        // direct url to the thumbnail of the show provided by the provider
	Detailed      bool          // False means some details can be requested
	DRM           bool          // True when video is protected with DRM
	ShowPitch     string        // Pitch for the whole serie
	ShowURL       string        // url to the show page at provider
	ShowBannerURL string        // Banner for the show
	Category      string        // Show's category
	Provider      string        // provider's name
	Destination   string        // Destination code taken from watch list
}
