package gulli

import (
	"io"
	"log"
	"net/http"
	"path/filepath"

	tvhttp "github.com/simulot/aspiratv/net/http"
	"github.com/simulot/aspiratv/parsers/htmlparser"
	"github.com/simulot/aspiratv/providers"
)

type getter interface {
	Get(uri string) (io.Reader, error)
}

type Gulli struct {
	getter            getter
	htmlParserFactory *htmlparser.Factory
	seenShows         map[string]bool
	debug             bool
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
		getter:            tvhttp.DefaultClient,
		htmlParserFactory: nil,
		seenShows:         map[string]bool{},
	}
	for _, f := range conf {
		f(p)
	}
	if rt, ok := p.getter.(http.RoundTripper); ok {
		p.htmlParserFactory = htmlparser.NewFactory(htmlparser.SetTransport(rt))
	} else {
		p.htmlParserFactory = htmlparser.NewFactory()
	}

	return p, nil
}

// SetDebug set debug mode
func (p *Gulli) SetDebug(b bool) {
	p.debug = b
}

// withGetter set a getter for Gulli
func withGetter(g getter) func(p *Gulli) {
	return func(p *Gulli) {
		p.getter = g
	}
}

// Name return the name of the provider
func (p Gulli) Name() string { return "gulli" }

// Shows download the shows catalog from the web site.
func (p *Gulli) Shows(mm []*providers.MatchRequest) ([]*providers.Show, error) {
	shows := []*providers.Show{}
	log.Print("[gulli] Fetch Gulli's new shows")
	shows, err := p.getAll(mm)
	if err != nil {
		return nil, err
	}
	return shows, err
}

// GetShowStreamURL return the show's URL, a mp4 file
func (p *Gulli) GetShowStreamURL(s *providers.Show) (string, error) {
	return s.StreamURL, nil
}

// GetShowInfo gather show information from dedicated web page.
// It load the html page of the show to extract availability date used as airdate and production year as season
func (p *Gulli) GetShowInfo(s *providers.Show) error {
	return nil
}

// GetShowFileName return a file name with a path that is compatible with PLEX server:
//   ShowName/Season NN/ShowName - sNNeMM - Episode title
//   Show and Episode names are sanitized to avoid problem when saving on the file system
func (p *Gulli) GetShowFileName(s *providers.Show) string {
	return filepath.Join(
		providers.PathNameCleaner(s.Show),
		"Season "+providers.Format2Digits(s.Season),
		providers.FileNameCleaner(s.Show)+" - s"+providers.Format2Digits(s.Season)+"e"+providers.Format2Digits(s.Episode)+" - "+providers.FileNameCleaner(s.Title)+".mp4",
	)
}

// GetShowFileNameMatcher return a file pattern of this show
// used for detecting already got episode even when episode or season is different
func (Gulli) GetShowFileNameMatcher(s *providers.Show) string {
	return filepath.Join(
		providers.PathNameCleaner(s.Show),
		"*",
		providers.FileNameCleaner(s.Show)+" - * - "+providers.FileNameCleaner(s.Title)+".mp4",
	)
}
