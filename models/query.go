package models

import (
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type SearchResult struct {
	Err         error     // When not nil, the rest of the structure is invalid
	ID          string    // To disambiguate detail search
	Type        MediaType // Collection / Series / TV Show / Movie or media
	Title       string    // Title
	Plot        string    // Plot
	PageURL     string    // Page on the web site
	ThumbURL    string    // Image
	AvailableOn time.Time // O when available, or date of availability
	Chanel      string    // TV Chanel
	Provider    string    // Provider
	IsPlayable  bool      // True when available for free streaming from TV site
	IsTeaser    bool      // True when only a teaser is available
}

type SearchQuery struct {
	Title           string // Title to be searched on line
	normalizedTitle string // Normalized title to ease comparisons with diacritics

	// Future Use
	// MustTitle       *regexp.Regexp // When given, the title must be recognized by this regexp
	// RejectTitle     *regexp.Regexp // When give the title must not be recognized by this regexp
	// AiredAfter      time.Time      // WHen set, must be aired after
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
