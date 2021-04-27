package models

import (
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var MediaTypeLabel = []string{
	"--?--",      // TypeUnknown
	"Collection", // TypeCollection           // We don't know yet the actual type, but this is a collection, Arte?
	"Série",      // TypeSeries               // Series with seasons and episodes
	"Magazine",   // TypeTVShow               // TV Show or magazine
	"Média",      // TypeMovie                // Movie
}

type SearchResult struct {
	Err             error     // When not nil, the rest of the structure is invalid
	ID              string    // To disambiguate detail search
	Rank            int       // Result rank
	Type            MediaType // Collection / Series / TV Show / Movie or media
	Show            string    // Show name
	Title           string    // Media Title
	Plot            string    // Plot
	PageURL         string    // Page on the web site
	ThumbURL        string    // Image
	Aired           time.Time // When it has been broadcasted
	AvailableOn     time.Time // O when available, or date of availability
	Chanel          string    // TV Chanel
	Provider        string    // Provider
	Tags            []string
	AvailableVideos int
	MoreAvailable   bool
}

type AvailableSeason struct {
	Season             int
	AvailableEpsisodes int
	LatestAired        time.Time
}

func (sr *SearchResult) AddTag(t string) {
	for _, s := range sr.Tags {
		if s == t {
			return
		}
	}
	sr.Tags = append(sr.Tags, t)
}

type SearchQuery struct {
	Title           string    // Title to be searched on line
	normalizedTitle string    // Normalized title to ease comparisons with diacritics
	OnlyExactTitle  bool      // When true, the result title must match the searched title
	AiredAfter      time.Time // WHen set, must be aired after, zero means during the last month

	// Future Use
	// MustTitle       *regexp.Regexp // When given, the title must be recognized by this regexp
	// RejectTitle     *regexp.Regexp // When give the title must not be recognized by this regexp
	// AiredBefore     time.Time      // when set, must be aired before

	// // Default options
	// RefuseSeries  bool
	// RefuseTVShow  bool
	// RefuseMovies  bool
	// AcceptBonuses bool
}

func (q SearchQuery) IsMatch(t string) bool {

	// if q.MustTitle != nil {
	// 	if !q.MustTitle.MatchString(q.Title) {
	// 		return false
	// 	}
	// }
	// if q.RejectTitle != nil {
	// 	if !q.RejectTitle.MatchString(q.Title) {
	// 		return false
	// 	}
	// }
	// if q.MustTitle != nil && q.RejectTitle != nil {
	// 	return true
	// }

	if len(q.normalizedTitle) == 0 {
		q.normalizedTitle = normalize(q.Title)
	}

	normalizedT := normalize(t)
	return strings.Contains(normalizedT, q.normalizedTitle)
}

// Transform strings with diacritics into comparable strings
// https://stackoverflow.com/questions/26722450/remove-diacritics-using-go
func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

var removeSpaces = regexp.MustCompile(`[ ]{2,}`)

func normalize(s string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	s, _, _ = transform.String(t, s)

	return removeSpaces.ReplaceAllString(strings.TrimSpace(strings.ToLower(s)), " ")
}

/*
type SubscriptionInterface interface {
	GetSubscription(id string) (Subscription, error)
	SetSubscription(s Subscription) (Subscription, error)
	GetSubciptionList() ([]Subscription, error)
}

type Subscription struct {
}
*/
