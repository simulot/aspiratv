package gulli

import (
	"fmt"
	"io"
	"log"

	"github.com/simulot/aspiratv/net/http"
	"github.com/simulot/aspiratv/parsers/htmlparser"
	"github.com/simulot/aspiratv/providers"
)

const (
	gulliAll = "http://replay.gulli.fr/all"
)

type getter interface {
	Get(uri string) (io.Reader, error)
}

type Gulli struct {
	getter getter
	parser *htmlparser.Factory
}

// init registers Gulli provider
func init() {
	p, err := New()
	if err != nil {
		panic(err)
	}
	providers.Register(p)
}

func New(conf ...func(p *Gulli)) (*Gulli, error) {
	p := &Gulli{
		getter: http.DefaultClient,
		parser: htmlparser.NewFactory(),
	}
	return p, nil
}

// Name return the name of the provider
func (p Gulli) Name() string { return "gulli" }

// Shows download the shows catalog from the web site.
func (p *Gulli) Shows(mm []*providers.MatchRequest) ([]*providers.Show, error) {
	shows := []*providers.Show{}
	log.Print("[gulli] Fetch Gulli's new shows")
	gss, err := getAllShowList(p.parser.New(), gulliAll)
	if err != nil {
		return nil, err
	}
	for _, gs := range gss {
		s := &providers.Show{
			Show:         gs.name,
			ThumbnailURL: gs.thumbnail,
			ShowURL:      gs.url,
		}
		if providers.IsShowMatch(mm, s) {
			shows = append(shows, s)
		}

	}
	return nil, fmt.Errorf("[gulli] Shows not implented")
}

// GetShowStreamURL return the show's URL, a mp4 file
func (p *Gulli) GetShowStreamURL(s *providers.Show) (string, error) {
	return "", fmt.Errorf("[gulli] GetShowStreamURL not implented")
}

// GetShowInfo gather show information from dedicated web page.
// It load the html page of the show to extract availability date used as airdate and production year as season
func (p *Gulli) GetShowInfo(s *providers.Show) error {
	return fmt.Errorf("[gulli] GetShowInfo not implented")
}

// GetShowFileName return a file name with a path that is compatible with PLEX server:
//   ShowName/Season NN/ShowName - sNNeMM - Episode title
//   Show and Episode names are sanitized to avoid problem when saving on the file system
func (p *Gulli) GetShowFileName(s *providers.Show) string {
	return ""
}

// GetShowFileNameMatcher return a file pattern of this show
// used for detecting already got episode even when episode or season is different
func (Gulli) GetShowFileNameMatcher(s *providers.Show) string {
	return ""
}
