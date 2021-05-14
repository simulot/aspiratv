package library

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/simulot/aspiratv/pkg/models"
)

var RegularNameCleaner = NewNameCleaner(
	// illegal chars in windows filename  / \ : * < > ? "
	strings.NewReplacer(
		"/", "-",
		"\\", "-",
		"?", "",
		":", "-",
		"*", "-",
		"|", "-",
		"\"", "''",
		">", ")",
		"<", "(",
	))

var UTF8NameCleaner = NewNameCleaner(
	strings.NewReplacer(
		"/", "\u2215", // ∕
		"\\", "\u2216", // ∖
		"?", "\uFF1F", // ？
		":", "\u02F8", // ˸
		"*", "\u2731", // ✱
		"|", "\u2758", // ❘
		"\"", "\uff02", // ＂
		"<", "\uFF1C", // ＜
		">", "\uFF1E", // ＞
	))

var DefaultPathSettings = map[models.MediaType]models.PathSettings{
	models.TypeCollection: {
		ShowPathTemplate:      `{{.Show}}`,
		SeasonPathTemplate:    ``,
		MediaFileNameTemplate: `{{.Title}}.mp4`,
	},
	models.TypeSeries: {
		ShowPathTemplate:      `{{.Show}}`,
		SeasonPathTemplate:    `{{if not .IsBonus}}Season {{.Season | printf "%02d" }}{{else}}Specials{{end}}`,
		MediaFileNameTemplate: `{{if not .IsBonus}}{{.Show}} s{{.Season | printf "%02d" }}e{{.Episode | printf "%02d" }} {{end}}{{.Title}}.mp4`,
	},
	models.TypeTVShow: {
		ShowPathTemplate:      `{{.Show}}`,
		SeasonPathTemplate:    `{{if not .IsBonus}}Season {{.Year | printf "%04d" }}{{else}}Specials{{end}}`,
		MediaFileNameTemplate: `{{if not .IsBonus}}{{.Show}} {{.Aired.Format "2006-01-02" }} {{end}}{{.Title}}.mp4`,
	},
	models.TypeMovie: {
		ShowPathTemplate:      `{{.Title}} ({{.Year}})`,
		SeasonPathTemplate:    ``,
		MediaFileNameTemplate: `{{.Title}} ({{.Year}}).mp4`,
	},
}
var DefaultFileNamer = map[models.MediaType]*FilesNamer{
	models.TypeCollection: MustFileName(DefaultPathSettings[models.TypeCollection], RegularNameCleaner),
	models.TypeSeries:     MustFileName(DefaultPathSettings[models.TypeSeries], RegularNameCleaner),
	models.TypeTVShow:     MustFileName(DefaultPathSettings[models.TypeTVShow], RegularNameCleaner),
	models.TypeMovie:      MustFileName(DefaultPathSettings[models.TypeMovie], RegularNameCleaner),
}

type FilesNamer struct {
	nc                *NameCleaner
	tmplShowPath      *template.Template
	tmplSeasonPath    *template.Template
	tmplMediaFileName *template.Template
}

func NewFilesNamer(s models.PathSettings, nc *NameCleaner) (*FilesNamer, error) {
	var err error
	n := FilesNamer{
		nc: nc,
	}
	n.tmplShowPath, err = template.New("ShowPath").Parse(s.ShowPathTemplate)
	if err != nil {
		return nil, fmt.Errorf("invalid ShowPath template: %w", err)
	}
	n.tmplSeasonPath, err = template.New("SeasonPath").Parse(s.SeasonPathTemplate)
	if err != nil {
		return nil, fmt.Errorf("invalid SeasonPath template: %w", err)
	}
	n.tmplMediaFileName, err = template.New("MediaFileName").Parse(s.MediaFileNameTemplate)
	if err != nil {
		return nil, fmt.Errorf("invalid MediaFileName template: %w", err)
	}
	return &n, nil
}

func MustFileName(s models.PathSettings, nc *NameCleaner) *FilesNamer {
	n, err := NewFilesNamer(s, nc)
	if err != nil {
		panic(err)
	}
	return n
}

func (n FilesNamer) ShowPath(info models.MediaInfo) (string, error) {
	var err error
	b := bytes.NewBuffer(nil)
	err = n.tmplShowPath.Execute(b, info)
	if err != nil {
		return "", err
	}
	return n.nc.Clean(b.String()), nil
}

func (n FilesNamer) SeasonPath(info models.MediaInfo) (string, error) {
	var err error
	b := bytes.NewBuffer(nil)
	err = n.tmplSeasonPath.Execute(b, info)
	if err != nil {
		return "", err
	}
	return n.nc.Clean(b.String()), nil
}

func (n *FilesNamer) MediaFileName(info models.MediaInfo) (string, error) {
	var err error
	b := bytes.NewBuffer(nil)
	err = n.tmplMediaFileName.Execute(b, info)
	if err != nil {
		return "", err
	}

	name := b.String()
	ext := filepath.Ext(name)
	base := n.nc.Clean(strings.TrimSuffix(name, ext))

	ext = strings.TrimSpace(anySpace.ReplaceAllString(ext, ""))

	return base + ext, nil
}

type NameCleaner struct {
	repl *strings.Replacer
}

var (
	doubleSpace = regexp.MustCompile(`[ ]{2,}`)
	anySpace    = regexp.MustCompile(`[ ]+`)
)

func NewNameCleaner(repl *strings.Replacer) *NameCleaner {
	return &NameCleaner{
		repl: repl,
	}
}

func (n NameCleaner) Clean(s string) string {
	if n.repl != nil {
		s = n.repl.Replace(s)
	}
	s = doubleSpace.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	return s
}
