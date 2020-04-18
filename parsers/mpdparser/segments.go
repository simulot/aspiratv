package mpdparser

type segmentVariable struct {
	Number           int    // $Number$
	RepresentationID string // $RepresentationID$
	Duration         int
	StartNumber      int
	TimeScale        int
}

func (p *MPDParser) Segements(a *AdaptationSet, ID string) chan string {

	if len(a.SegmentTemplate.SegmentTimeline.S) > 0 {
		return segmentTimeLine(a)
	}

	return nil
}

func segmentTimeLine(a *AdaptationSet, ID string) chan string {

}
