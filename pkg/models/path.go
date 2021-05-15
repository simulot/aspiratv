package models

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var DefaultPathSettings = map[PathNamingType]PathSettings{
	PathTypeCollection: {
		ShowPathTemplate:      `{{.Show}}`,
		SeasonPathTemplate:    ``,
		MediaFileNameTemplate: `{{.Title}}.mp4`,
	},
	PathTypeSeries: {
		ShowPathTemplate:      `{{.Show}}`,
		SeasonPathTemplate:    `{{if not .IsBonus}}Season {{.Season | printf "%02d" }}{{else}}Specials{{end}}`,
		MediaFileNameTemplate: `{{if not .IsBonus}}{{.Show}} s{{.Season | printf "%02d" }}e{{.Episode | printf "%02d" }} {{end}}{{.Title}}.mp4`,
	},
	PathTypeTVShow: {
		ShowPathTemplate:      `{{.Show}}`,
		SeasonPathTemplate:    `{{if not .IsBonus}}Season {{.Year | printf "%04d" }}{{else}}Specials{{end}}`,
		MediaFileNameTemplate: `{{if not .IsBonus}}{{.Show}} {{.Aired.Format "2006-01-02" }} {{end}}{{.Title}}.mp4`,
	},
	PathTypeMovie: {
		ShowPathTemplate:      `{{.Title}} ({{.Year}})`,
		SeasonPathTemplate:    ``,
		MediaFileNameTemplate: `{{.Title}} ({{.Year}}).mp4`,
	},
}
var DefaultFileNamer = map[PathNamingType]*FileNamer{
	PathTypeCollection: MustFileName(DefaultPathSettings[PathTypeCollection], RegularNameCleaner),
	PathTypeSeries:     MustFileName(DefaultPathSettings[PathTypeSeries], RegularNameCleaner),
	PathTypeTVShow:     MustFileName(DefaultPathSettings[PathTypeTVShow], RegularNameCleaner),
	PathTypeMovie:      MustFileName(DefaultPathSettings[PathTypeMovie], RegularNameCleaner),
}

// FileNamer build templates to be used when creating files/folders names
type FileNamer struct {
	nc                *NameCleaner
	tmplShowPath      *template.Template
	tmplSeasonPath    *template.Template
	tmplMediaFileName *template.Template
}

func NewFilesNamer(s PathSettings, nc *NameCleaner) (*FileNamer, error) {
	var err error
	n := FileNamer{
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

func MustFileName(s PathSettings, nc *NameCleaner) *FileNamer {
	n, err := NewFilesNamer(s, nc)
	if err != nil {
		panic(err)
	}
	return n
}

func (n FileNamer) ShowPath(info MediaInfo) (string, error) {
	var err error
	b := bytes.NewBuffer(nil)
	err = n.tmplShowPath.Execute(b, info)
	if err != nil {
		return "", err
	}
	return n.nc.Clean(b.String()), nil
}

func (n FileNamer) SeasonPath(info MediaInfo) (string, error) {
	var err error
	b := bytes.NewBuffer(nil)
	err = n.tmplSeasonPath.Execute(b, info)
	if err != nil {
		return "", err
	}
	return n.nc.Clean(b.String()), nil
}

func (n *FileNamer) MediaFileName(info MediaInfo) (string, error) {
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

// RegularNameCleaner is used to avoid forbiedden file names created after show title
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

// UTF8NameCleaner replace forbidden chars by some UTF-8 with similar glyph **EXPERIMENTAL**
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

// Clean file name from forbidden chars, multiples spaces and trailing spaces
func (n NameCleaner) Clean(s string) string {
	if n.repl != nil {
		s = n.repl.Replace(s)
	}
	s = doubleSpace.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	return s
}
