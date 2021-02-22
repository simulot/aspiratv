package download

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/simulot/aspiratv/matcher"
	"github.com/simulot/aspiratv/metadata/nfo"
)

var (
	fileNameReplacer = strings.NewReplacer("/", "-", "\\", "-", "!", "", "?", "", ":", "-", ",", "", "*", "-", "|", "-", "\"", "", ">", "", "<", "")
	pathNameReplacer = strings.NewReplacer("!", "", "?", "", ":", " ", ",", " ", "*", "", "|", " ", "\"", "", ">", "", "<", "", " - ", " ")
)

// FileNameCleaner return a safe file name from a given show name.
func FileNameCleaner(s string) string {
	return strings.TrimSpace(fileNameReplacer.Replace(s))
}

// PathNameCleaner return a safe path name from a given show name.
func PathNameCleaner(s string) string {
	if i := strings.Index(s, ":"); i >= 0 && i < 2 {
		return s[:i] + strings.TrimSpace(pathNameReplacer.Replace(s[i:]))
	}
	return strings.TrimSpace(pathNameReplacer.Replace(s))
}

// Format2Digits return a number with 2 digits when there is only one digit
func Format2Digits(d string) string {
	if len(d) < 2 {
		return "0" + d
	}
	return d
}

// Plex naming convention
//    https://support.plex.tv/articles/naming-and-organizing-your-tv-show-files/

var (
	seasonTemplates = map[nfo.ShowType]*template.Template{
		nfo.TypeShow:   template.Must(template.New("serieTVShow").Parse(`{{if not .IsBonus}}Season {{.Aired.Time.Year | printf "%04d" }}{{else}}Specials{{end}}`)),
		nfo.TypeSeries: template.Must(template.New("serieSeason").Parse(`{{if not .IsBonus}}Season {{.Season | printf "%02d" }}{{else}}Specials{{end}}`)),
		nfo.TypeMovie:  nil,
	}
	showNameTemplates = map[nfo.ShowType]*template.Template{
		nfo.TypeShow:   template.Must(template.New("serieTVShow").Parse(`{{.Showtitle}} - {{.Aired.Time.Format "2006-01-02"}}.mp4`)),
		nfo.TypeSeries: template.Must(template.New("serieShowName").Parse(`{{.Showtitle}} - s{{.Season | printf "%02d" }}e{{.Episode | printf "%02d" }} - {{.Title}}.mp4`)),
		nfo.TypeMovie:  template.Must(template.New("movieName").Parse(`{{.Title}}.mp4`)),
	}
)

// SeasonPath returns the full path for the episode's season according the template
func SeasonPath(showPath string, m *matcher.MatchRequest, info *nfo.MediaInfo) (string, error) {
	var err error

	seasonPart := &strings.Builder{}
	t := info.MediaType
	if t == nfo.TypeNotSpecified {
		t = nfo.TypeSeries
	}

	seasonTmpl := seasonTemplates[t]
	if m.SeasonPathTemplate != nil && m.SeasonPathTemplate.T != nil {
		seasonTmpl = m.SeasonPathTemplate.T
	}

	if seasonTmpl != nil {
		err = seasonTmpl.Execute(seasonPart, info)
		if err != nil {
			return "", fmt.Errorf("Can't use this season template: %w", err)
		}
	}
	return filepath.Join(showPath, seasonPart.String()), nil
}

// MediaPath returns the full path for an episode using filename template when present
func MediaPath(showPath string, m *matcher.MatchRequest, info *nfo.MediaInfo) (string, error) {
	var err error

	seasonPart, err := SeasonPath(showPath, m, info)
	if err != nil {
		return "", err
	}

	t := info.MediaType
	showPart := &strings.Builder{}
	showTmpl := showNameTemplates[t]
	if m.ShowNameTemplate != nil && m.ShowNameTemplate.T != nil {
		showTmpl = m.ShowNameTemplate.T
	}

	err = showTmpl.Execute(showPart, info)
	if err != nil {
		return "", fmt.Errorf("Can't use this name template: %w", err)
	}

	return filepath.Join(seasonPart, FileNameCleaner(showPart.String())), nil
}
