package francetv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/gocolly/colly"
	"github.com/simulot/aspiratv/media"
	"github.com/simulot/aspiratv/metadata/nfo"
	"github.com/simulot/aspiratv/net/myhttp/httptest"
)

type FTVPlayerVideo struct {
	ContentID int    `json:"contentId"`
	VideoID   string `json:"videoId"`
	// EndDate   time.Time `json:"endDate"`
	// OriginURL string    `json:"originUrl"`
	// // ComingNext         ComingNext         `json:"comingNext"`
	// IsSponsored   bool        `json:"isSponsored"`
	// IsAdVisible   interface{} `json:"isAdVisible"`
	// VideoTitle    string      `json:"videoTitle"`
	// ProgramName   string      `json:"programName"`
	// SeasonNumber  int         `json:"seasonNumber"`
	// EpisodeNumber int         `json:"episodeNumber"`
	// LayerType     interface{} `json:"layerType"`
	// RatingCsaCode string      `json:"ratingCsaCode"`
	// Logo          interface{} `json:"logo"`
	// Name          interface{} `json:"name"`
	// // BroadcastBeginDate BroadcastBeginDate `json:"broadcastBeginDate"`
	// Intro bool `json:"intro"`
}

// GetMediaDetails download more details when available  especially the stream URL.
// The player webservice returns some metadata and an URL named Token.
// The must been acquired right before the actual download. It has a limited validity
// In the structure returned by token URL, another URL is provided. The request is then redirected
// to the actual video stream. This url has also a limited validity.
//
// But for some reason FFMPEG doesn't follow the redirection. So, we have to get the final URL before
// calling FFMPEG // FranceTV provides a subtitle tracks that isn't decoded by FFMPEG.
// And FFMPEG doesn't get always  the best video resolution
//
// The video stream is in fact a MPD manifest files. We can edit this manifest for removing unwanted tracks.
//

func (p *FranceTV) GetMediaDetails(ctx context.Context, m *media.Media) error {
	info := m.Metadata.GetMediaInfo()
	parser := p.htmlParserFactory.New() // TODO withContext
	videoID := ""

	parser.OnHTML("meta", func(e *colly.HTMLElement) {
		switch e.Attr("property") {
		case "og:image":
			info.Thumb = append(info.Thumb, nfo.Thumb{
				URL: e.Attr("content"),
			})
		case "og:description":
			info.Plot = e.Attr("content")
		case "video:actor":
			for _, a := range strings.Split(e.Attr("content"), ",") {
				info.Actor = append(info.Actor, nfo.Actor{Name: strings.TrimSpace(a)})
			}
		case "video:director":
			for _, a := range strings.Split(e.Attr("content"), ",") {
				info.Director = append(info.Director, a)
			}
			// TODO: get exact aired time see https://github.com/simulot/aspiratv/issues/63#issuecomment-779027415
			// case "video:release_date":
			// 	t, _ := time.Parse("2006-01-02T15:04:05-07:00", e.Attr("content"))
			// 	info.Aired = nfo.Aired(t)
		}

	})

	if info.Season == 0 && info.Episode == 0 {
		info.Season = info.Aired.Time().Year()
	}

	parser.OnHTML("div.l-content script", func(e *colly.HTMLElement) {
		if !strings.Contains(e.Text, "FTVPlayerVideos") {
			return
		}
		start := strings.Index(e.Text, "[")
		end := strings.Index(e.Text, "];")
		if start < 0 || end < 0 {
			return
		}

		s := e.Text[:end+1][start:]
		// videos := []FTVPlayerVideos{}
		var videos []FTVPlayerVideo

		err := json.Unmarshal([]byte(s), &videos)
		if err != nil {
			p.config.Log.Error().Printf("[%s] Can't decode FTVPlayerVideo: %s", p.Name(), err)
			return
		}
		videoID = videos[0].VideoID
	})
	err := parser.Visit(info.PageURL)
	if err != nil {
		return err
	}

	err = p.getMediaURL(ctx, info, videoID)
	return err
}

func (p *FranceTV) getMediaURL(ctx context.Context, info *nfo.MediaInfo, videoID string) error {
	v := url.Values{}
	v.Set("country_code", "FR")
	v.Set("w", "1920")
	v.Set("h", "1080")
	v.Set("version", "5.18.3")
	v.Set("domain", "www.france.tv")
	v.Set("device_type", "desktop")
	v.Set("browser", "firefox")
	v.Set("browser_version", "85")
	v.Set("os", "windows")
	v.Set("gmt", "+1")

	u := "https://player.webservices.francetelevisions.fr/v1/videos/" + videoID + "?" + v.Encode()
	p.config.Log.Debug().Printf("[%s] Player URL for title '%s' is %q.", p.Name(), info.Title, u)

	r, err := p.getter.Get(ctx, u)
	if err != nil {
		return fmt.Errorf("Can't get player: %w", err)
	}
	if p.config.Log.IsDebug() {
		r = httptest.DumpReaderToFile(p.config.Log, r, "francetv-player-"+videoID+"-")
	}
	defer r.Close()

	pl := player{}
	err = json.NewDecoder(r).Decode(&pl)
	if err != nil {
		return fmt.Errorf("Can't decode player: %w", err)
	}

	// Get Token
	if len(pl.Video.Token) > 0 {
		p.config.Log.Debug().Printf("[%s] Player token for '%s' is %q ", p.Name(), info.Title, pl.Video.Token)

		r2, err := p.getter.Get(ctx, pl.Video.Token)
		if err != nil {
			return fmt.Errorf("Can't get token %s: %w", pl.Video.Token, err)
		}
		if p.config.Log.IsDebug() {
			r2 = httptest.DumpReaderToFile(p.config.Log, r2, "francetv-token-"+videoID+"-")
		}
		defer r2.Close()
		pl := struct {
			URL string `json:"url"`
		}{}
		err = json.NewDecoder(r2).Decode(&pl)
		if err != nil {
			return fmt.Errorf("Can't decode token's url : %w", err)
		}
		if len(pl.URL) == 0 {
			return fmt.Errorf("Show's URL is empty")
		}
		info.MediaURL = pl.URL

	}
	p.config.Log.Trace().Printf("[%s] Player URL for '%s' is %q ", p.Name(), info.Title, info.MediaURL)
	return nil
}
