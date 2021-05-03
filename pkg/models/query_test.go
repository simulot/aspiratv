package models

import "testing"

func TestQuery(t *testing.T) {
	tc := []struct {
		q     SearchQuery
		title string
		want  bool
	}{
		{
			SearchQuery{
				Title: "my search",
			},
			"nope",
			false,
		},
		{
			SearchQuery{
				Title: "my search",
			},
			"my search",
			true,
		},
		{
			SearchQuery{
				Title: "my search",
			},
			"My Search",
			true,
		},
		{
			SearchQuery{
				Title: "my search",
			},
			"\tMy Search ",
			true,
		},
		{
			SearchQuery{
				Title: "francais",
			},
			"Français",
			true,
		},
		{
			SearchQuery{
				Title: "un                     café francais",
			},
			"Un Café Français",
			true,
		},
	}

	for _, c := range tc {
		t.Run(c.title, func(t *testing.T) {
			got := c.q.IsMatch(c.title)
			if c.want != got {
				t.Errorf("Expecting %v, got %v", c.want, got)
			}
		})
	}
}
