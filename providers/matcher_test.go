package providers

import "testing"

func TestIsShowMatch(t *testing.T) {
	ml := []*MatchRequest{
		{
			Show:        "searched1",
			Provider:    "p1",
			Destination: "dest1",
		},
		{
			Show:        "searched2",
			Provider:    "p2",
			Destination: "dest1",
		},
		{
			Show:        "searched3",
			Provider:    "p1",
			Destination: "dest2",
		},
	}
	type args struct {
		mm []*MatchRequest
		s  *Show
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Show not found",
			args{
				ml,
				&Show{
					Show:     "I'm not here",
					Provider: "p1",
				},
			},
			false,
		},
		{
			"Wrong provider",
			args{
				ml,
				&Show{
					Show:     "Searched1",
					Provider: "p2",
				},
			},
			false,
		},
		{
			"found Searched1",
			args{
				ml,
				&Show{
					Show:     "Searched1",
					Provider: "p1",
				},
			},
			true,
		},
		{
			"found Searched3",
			args{
				ml,
				&Show{
					Show:     "Searched3",
					Provider: "p1",
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsShowMatch(tt.args.mm, tt.args.s); got != tt.want {
				t.Errorf("IsShowMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
