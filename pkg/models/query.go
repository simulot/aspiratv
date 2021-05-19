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
	Err             error     `json:"err,omitempty"`          // When not nil, the rest of the structure is invalid
	ID              string    `json:"id,omitempty"`           // To disambiguate detail search
	Rank            int       `json:"rank,omitempty"`         // Result rank
	Type            MediaType `json:"type,omitempty"`         // Collection / Series / TV Show / Movie or media
	Show            string    `json:"show,omitempty"`         // Show name
	Title           string    `json:"title,omitempty"`        // Media Title
	Plot            string    `json:"plot,omitempty"`         // Plot
	PageURL         string    `json:"page_url,omitempty"`     // Page on the web site
	ThumbURL        string    `json:"thumb_url,omitempty"`    // Image
	Aired           time.Time `json:"aired,omitempty"`        // When it has been broadcasted
	AvailableOn     time.Time `json:"available_on,omitempty"` // O when available, or date of availability
	Chanel          string    `json:"chanel,omitempty"`       // TV Chanel
	Provider        string    `json:"provider,omitempty"`     // Provider
	Tags            []string  `json:"tags,omitempty"`
	AvailableVideos int       `json:"available_videos,omitempty"`
	MoreAvailable   bool      `json:"more_available,omitempty"`
}

type AvailableSeason struct {
	Season             int       `json:"season,omitempty"`
	AvailableEpsisodes int       `json:"available_epsisodes,omitempty"`
	LatestAired        time.Time `json:"latest_aired,omitempty"`
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
	Title           string    `json:"title,omitempty"`            // Title to be searched on line
	normalizedTitle string    `json:"-,omitempty"`                // Normalized title to ease comparisons with diacritics
	OnlyExactTitle  bool      `json:"only_exact_title,omitempty"` // When true, the result title must match the searched title
	AiredAfter      time.Time `json:"aired_after,omitempty"`      // WHen set, must be aired after, zero means during the last month
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
