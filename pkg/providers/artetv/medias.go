package artetv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/simulot/aspiratv/pkg/download"
	"github.com/simulot/aspiratv/pkg/models"
	"github.com/simulot/aspiratv/pkg/myhttp"
)

func (p *Arte) GetMedias(ctx context.Context, task models.DownloadTask) (<-chan models.DownloadItem, error) {
	c := make(chan models.DownloadItem, 1)
	go func() {
		defer close(c)
		// Read collection / show page
		parser := colly.NewCollector()
		pgm := InitialProgram{}
		js := ""
		parser.OnHTML("body > script", func(e *colly.HTMLElement) {

			// Get JSON with collection data from the HTML page of the collection
			start := strings.Index(e.Text, "{")
			end := strings.LastIndex(e.Text, "}")
			if start < 0 || end < 0 {
				return
			}

			js = e.Text[:end+1][start:]
		})
		err := p.limiter.Wait(ctx)
		if err != nil {
			log.Printf("[ARTETV] Decoding program's page: %s", err)
			return
		}
		err = parser.Visit(task.Result.PageURL)
		if err != nil {
			log.Printf("[ARTETV] Decoding program's page: %s", err)
			return
		}

		err = json.NewDecoder(strings.NewReader(js)).Decode(&pgm)
		if err != nil {
			log.Printf("[ARTETV] Decoding program's page: %s", err)
			return
		}

		for _, page := range pgm.Pages.List {
			ShowInfo := models.ShowInfo{}
			for _, zone := range page.Zones {

				// Get show level info
				if zone.Code.Name == "collection_content" {
					for _, data := range zone.Data {
						ShowInfo.Title = data.Title
						ShowInfo.Plot = data.Description
						ShowInfo.Images = getImages(data.Images)
						ShowInfo.Type = task.Result.Type
					}
					continue
				}

				// Get medias from show page
				if zone.Code.Name == "collection_subcollection" || zone.Code.Name == "collection_videos" {
					for _, ep := range zone.Data {
						Info := models.MediaInfo{
							ID:       ep.ProgramID,
							Provider: "arte",
							Channel:  "Arte",
							Title:    getFirstString([]string{ep.Subtitle, ep.Title, ShowInfo.Title}),
							Show:     ShowInfo.Title,
							ShowInfo: &ShowInfo,
							Aired:    ep.Availability.Start.Time(),
							Year:     ep.Availability.Start.Time().Year(),
							PageURL:  ep.URL,
							IsBonus:  ep.Kind.Code == "BONNUS",
							Type:     task.Result.Type,
						}

						err = p.fetchShowDetails(ctx, &Info)
						if err != nil {
							log.Printf("[ARTETV] Can't get details: %s", err)
							continue
						}

						// Adjust showinfo
						ShowInfo.Type = Info.Type

						dl := models.DownloadItem{
							Downloader: download.NewFFMPEG().Input(Info.StreamURL),
							MediaInfo:  Info,
						}
						select {
						case <-ctx.Done():
							log.Printf("[ARTETV] %s", ctx.Err())
							return
						case c <- dl:
						}
					}
				}
			}
		}

	}()
	return c, nil
}

func (p *Arte) fetchShowDetails(ctx context.Context, info *models.MediaInfo) error {
	parser := colly.NewCollector()

	pgm := InitialProgram{}
	js := ""

	parser.OnHTML("body > script", func(e *colly.HTMLElement) {
		if !strings.Contains(e.Text, "__INITIAL_STATE__") {
			return
		}
		start := strings.Index(e.Text, "{")
		end := strings.LastIndex(e.Text, "}")
		if start < 0 || end < 0 {
			return
		}
		js = e.Text[:end+1][start:]
	})

	err := p.limiter.Wait(ctx)
	if err != nil {
		return err
	}
	err = parser.Visit(info.PageURL)
	if err != nil {
		return err
	}

	{
		b := myhttp.JSONDumper([]byte(js))
		w, err := os.Create("tmp/arte_" + info.ID + ".json")
		if err == nil {
			w.Write(b)
			w.Close()
		}
	}

	err = json.Unmarshal([]byte(js), &pgm)
	if err != nil {
		return err
	}
	player := p.parseMetaData(info, pgm)
	err = p.parsePlayer(ctx, info, player)
	if err != nil {
		return err
	}

	return nil
}

func (p *Arte) parseMetaData(info *models.MediaInfo, pgm InitialProgram) (player string) {
	var slug string
	for _, page := range pgm.Pages.List {
		slug = page.Slug

		p := &page.Parent
		for p != nil {
			info.Tags = append(info.Tags, p.Label)
			p = p.Parent
		}

		for _, zone := range page.Zones {
			if zone.Code.Name == "program_content" {

				for _, ep := range zone.Data {
					info.Images = getImages(ep.Images)
					info.Plot = ep.FullDescription                             // HTML?
					player = strings.Replace(ep.Player.Config, "v2", "v1", -1) // TODO use v2
					// Parse credits
					for _, credit := range ep.Credits {
						switch credit.Code {
						case "ACT":
							regActors := regexp.MustCompile(`^(.+)(?:\s\((.+)\))$|(.+)$`)
							for _, v := range credit.Values {
								actor := models.Person{}
								m := regActors.FindAllStringSubmatch(v, -1)
								if len(m) > 0 {
									if len(m[0]) == 4 {
										if len(m[0][3]) > 0 {
											actor.FullName = m[0][3]
										} else {
											actor.FullName = m[0][1]
											actor.Role = m[0][2]
										}
									}
								}
								info.Actors = append(info.Actors, actor)
							}
						case "PRODUCTION_YEAR":
							for _, v := range credit.Values {
								info.Year, _ = strconv.Atoi(v)
							}
						default:
							for _, v := range credit.Values {
								info.Credits = append(info.Credits, fmt.Sprintf("%s (%s)", v, credit.Label))
							}
						}
					}
				}
			}
		}
	}

	// Parse slug to extract episode number
	m := parseSlug.FindAllStringSubmatch(slug, -1)
	if len(m) > 0 && len(m[0]) > 0 {
		info.Season, _ = strconv.Atoi(m[0][1])
		info.Episode, _ = strconv.Atoi(m[0][2])
		info.Type = models.TypeSeries
		if info.Season == 0 {
			info.Season = 1
		}
	}
	return
}

func (p *Arte) parsePlayer(ctx context.Context, info *models.MediaInfo, playerRUL string) error {
	player := playerAPI{}
	req, err := p.client.NewRequestJSON(ctx, playerRUL, nil, nil, nil)
	if err != nil {
		return err
	}
	err = p.limiter.Wait(ctx)
	if err != nil {
		return err
	}
	err = p.client.GetJSON(req, &player)
	if err != nil {
		return err
	}

	info.StreamURL, err = p.getBestVideo(player.VideoJSONPlayer.VSR)
	if err != nil {
		return err
	}
	return nil
}

// getBestVideo return the best video stream given preferences
//   Streams are scored in following order:
//   - Version (VF,VF_ST) that match preference
//   - Stream quality, the highest possible

func (p *Arte) getBestVideo(ss map[string]StreamInfo) (string, error) {
	for _, v := range p.settings.PreferredVersions {
		for _, r := range p.preferredQuality {
			for _, s := range ss {
				if s.Quality == r && s.VersionCode == v {
					return s.URL, nil
				}
			}
		}
	}
	return "", errors.New("Can't find a suitable video stream")
}

func getImages(images map[string]Image) []models.Image {
	thumbs := []models.Image{}
	for k, i := range images {
		if len(i.BlurURL) == 0 {
			continue
		}
		aspect := "thumb"

		switch k {
		case "landscape":
			aspect = "thumb"
		case "banner":
			aspect = "backdrop"
		case "portrait":
			aspect = "poster"
		case "square":
			aspect = "poster"
		}

		thumbs = append(thumbs, models.Image{
			Aspect: aspect,
			URL:    getBestImage(i),
		})

	}
	return thumbs
}

// getBestImage retreive the url for the image of type "protrait/banner/landscape..." with the highest resolution
func getBestImage(image Image) string {
	bestResolution := 0
	bestURL := ""
	for _, r := range image.Resolutions {
		_ = 1
		res := r.H * r.W
		if res > bestResolution {
			bestURL = r.URL
			bestResolution = res
		}
	}
	if bestResolution == 0 {
		return ""
	}
	return bestURL
}

//https://www.arte.tv/guide/api/emac/v3/fr/web/programs/044892-008-A/?
//https://    www.arte.tv/guide/api/emac/v3/fr/web/data/COLLECTION_VIDEOS/?collectionId=RC-014408&page=1&limit=100
//https://api-cdn.arte.tv/      api/emac/v3/fr/web/data/COLLECTION_VIDEOS/?collectionId=RC-015842&page=2&limit=12
var (
	parseSlug       = regexp.MustCompile(`(\d+)?-(\d+)-(\d+)$`)       // Season, Episode, number of episodes in season
	parseSeason     = regexp.MustCompile(`Saison (\d+)`)              // Detect season number in web page
	parseEpisode    = regexp.MustCompile(`^(?:.+) \((\d+)\/(\d+)\)$`) // Extract episode number
	parseShowSeason = regexp.MustCompile(`^(.+) - Saison (\d+)$`)
)

func getFirstString(ss []string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}
