package library

import (
	"testing"
	"time"

	"github.com/simulot/aspiratv/pkg/models"
)

func TestMakeFileNameReplacer(t *testing.T) {
	tt := []struct {
		name     string
		expected string
		nc       *models.NameCleaner
	}{
		{"super simple regular", "super simple regular", models.RegularNameCleaner},
		{"super simple utf-8", "super simple utf-8", models.UTF8NameCleaner},
		{"game of thrones? <1/8> regular", "game of thrones (1-8) regular", models.RegularNameCleaner},
		{"game of thrones? <1/8> utf-8", "game of thrones\uFF1F \uFF1C1\u22158\uFF1E utf-8", models.UTF8NameCleaner},
	}

	for _, c := range tt {
		t.Run(c.name, func(t *testing.T) {
			if got := c.nc.Clean(c.name); got != c.expected {
				t.Errorf("Expecting %q, got %q", c.expected, got)
			}
		})
	}
}

func TestMakeFileNameSpacesManagement(t *testing.T) {
	tt := []struct {
		name     string
		expected string
		nc       *models.NameCleaner
	}{
		{"super simple", "super simple", models.RegularNameCleaner},
		{"double  spaces", "double spaces", models.RegularNameCleaner},
		{"triple   spaces", "triple spaces", models.RegularNameCleaner},
		{" space at ends ", "space at ends", models.RegularNameCleaner},
		{"    spaces at ends    ", "spaces at ends", models.RegularNameCleaner},
	}

	for _, c := range tt {
		t.Run(c.name, func(t *testing.T) {
			if got := c.nc.Clean(c.name); got != c.expected {
				t.Errorf("Expecting %q, got %q", c.expected, got)
			}
		})
	}
}

func TestFileNamer(t *testing.T) {
	cases := []struct {
		info models.MediaInfo
		showPath,
		seasonPath,
		mediaFile string
	}{
		{
			info: models.MediaInfo{
				Type:  models.TypeTVShow,
				Year:  2021,
				Aired: time.Date(2021, 05, 11, 0, 0, 0, 0, time.Local),
				Show:  "The Late Show with Stephen Colbert",
				Title: "Michelle Obama",
			},
			showPath:   "The Late Show with Stephen Colbert",
			seasonPath: "Season 2021",
			mediaFile:  "The Late Show with Stephen Colbert 2021-05-11 Michelle Obama.mp4",
		},
		{
			info: models.MediaInfo{
				Type:  models.TypeTVShow,
				Year:  2021,
				Aired: time.Date(2021, 05, 11, 0, 0, 0, 0, time.Local),
				Show:  "20 heures",
			},
			showPath:   "20 heures",
			seasonPath: "Season 2021",
			mediaFile:  "20 heures 2021-05-11.mp4",
		},
		{
			info: models.MediaInfo{
				Type:    models.TypeSeries,
				Year:    2021,
				Season:  1,
				Episode: 2,
				Aired:   time.Date(2021, 05, 16, 0, 0, 0, 0, time.Local),
				Show:    "The Nevers",
				Title:   "Exposure",
			},
			showPath:   "The Nevers",
			seasonPath: "Season 01",
			mediaFile:  "The Nevers s01e02 Exposure.mp4",
		},
		{
			info: models.MediaInfo{
				Type:    models.TypeSeries,
				Year:    2021,
				Aired:   time.Date(2021, 05, 16, 0, 0, 0, 0, time.Local),
				Show:    "The Nevers",
				Title:   "Making of",
				IsBonus: true,
			},
			showPath:   "The Nevers",
			seasonPath: "Specials",
			mediaFile:  "Making of.mp4",
		},
		{
			info: models.MediaInfo{
				Type:  models.TypeCollection,
				Aired: time.Date(2021, 03, 1, 0, 0, 0, 0, time.Local),
				Show:  "Rock Legends",
				Title: "The Clash",
			},
			showPath:   "Rock Legends",
			seasonPath: "",
			mediaFile:  "The Clash.mp4",
		},
		{
			info: models.MediaInfo{
				Type:    models.TypeSeries,
				Year:    2021,
				Season:  1,
				Episode: 4,
				Aired:   time.Date(2021, 05, 16, 0, 0, 0, 0, time.Local),
				Show:    "House of Cards",
				Title:   "Part 1/2",
			},
			showPath:   "House of Cards",
			seasonPath: "Season 01",
			mediaFile:  "House of Cards s01e04 Part 1-2.mp4",
		},
	}

	for _, c := range cases {
		t.Run(c.info.Show, func(t *testing.T) {
			got, err := models.DefaultFileNamer[models.PathNamingType(c.info.Type)].ShowPath(c.info)
			want := c.showPath
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
			}
			if c.showPath != got {
				t.Errorf("Expected ShowPath:%q, got: %q", want, got)
			}
			got, err = models.DefaultFileNamer[models.PathNamingType(c.info.Type)].SeasonPath(c.info)
			want = c.seasonPath
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
			}
			if c.seasonPath != got {
				t.Errorf("Expected SeasonPath:%q, got: %q", want, got)
			}
			got, err = models.DefaultFileNamer[models.PathNamingType(c.info.Type)].MediaFileName(c.info)
			want = c.mediaFile
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
			}
			if c.mediaFile != got {
				t.Errorf("Expected MediaFileName:%q, got: %q", want, got)
			}
		})

	}
}
