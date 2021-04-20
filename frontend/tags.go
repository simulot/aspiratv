package frontend

import (
	"fmt"
	"log"

	"github.com/maxence-charriere/go-app/v8/pkg/app"
)

type TagState int

const (
	TagUnselected TagState = iota
	TagSelected
	TagInactive
)

type TagInfo struct {
	// Tag's code, must be unique in tag list
	Code string

	// Tag's state Selected, Deselected,...
	State TagState

	// Tag' text
	Text string

	// Tag's count
	Count int

	// Icon
	Icon app.UI
}

type TagListOptions struct {
	// Non empty to enable All button at head of tags
	All *TagInfo

	// True if labels have a count
	CanCount bool

	// True if labels can be be disabled
	CanDisable bool

	// TODO: NotClickable
}

const allCode = "all*tags"

type TagList struct {
	app.Compo

	canDisable bool
	cantCount  bool
	allTag     *TagInfo

	// List of tags
	tags map[string]*TagInfo
}

func NewTagList(options *TagListOptions) *TagList {
	var opt TagListOptions
	if options != nil {
		opt = *options
	}

	l := TagList{
		canDisable: opt.CanDisable,
		cantCount:  opt.CanCount,
		tags:       map[string]*TagInfo{},
	}

	if opt.All != nil {
		l.allTag = opt.All
		l.allTag.Code = allCode
	}

	return &l
}

func (l *TagList) Update() {
	if l.Mounted() {
		l.Compo.Update()
	}
}

func (l *TagList) Render() app.UI {
	if len(l.tags) == 0 && l.allTag == nil {
		return nil
	}
	return app.Div().Body(
		app.If(l.allTag != nil, l.renderTag(l.allTag)),
		app.Range(l.tags).Map(func(k string) app.UI {
			return l.renderTag(l.tags[k])
		}),
	)
}

func (l *TagList) Reset() {
	if l.Mounted() {
		l.tags = map[string]*TagInfo{}
		if l.allTag != nil {
			l.allTag.State = TagSelected
			l.allTag.Count = 0
		}
		l.Update()
	}
}

func (l *TagList) GetTag(code string) *TagInfo {
	if code == allCode {
		return l.allTag
	}
	return l.tags[code]
}

func (l *TagList) GetState(code string) TagState {
	i := l.GetTag(code)
	if i == nil {
		return TagInactive
	}
	return i.State
}

func (l *TagList) SetTag(t *TagInfo) {
	if t.Code == allCode {
		if l.allTag == nil {
			return
		}
		l.allTag = t
		return
	}
	l.tags[t.Code] = t
	l.Update()
}

func (l *TagList) Toggle(code string) {
	t := l.GetTag(code)
	if t == nil {
		return
	}
	switch t.State {
	case TagSelected:
		t.State = TagUnselected
	case TagUnselected:
		if l.canDisable {
			t.State = TagInactive
		} else {
			t.State = TagSelected
		}
	case TagInactive:
		t.State = TagSelected
	}
	l.Update()
}

func (l *TagList) IncAll() int {
	if !l.cantCount || l.allTag == nil {
		return -1
	}
	l.allTag.Count++
	l.Update()
	return l.allTag.Count
}

func (l *TagList) SetOrIncTag(nt *TagInfo) int {
	if !l.cantCount {
		return -1
	}

	var t *TagInfo
	var ok bool
	if t, ok = l.tags[nt.Code]; !ok {
		t = nt
	}
	t.Count++
	l.tags[nt.Code] = t
	l.Update()
	return t.Count
}

func (l *TagList) renderTag(i *TagInfo) app.UI {
	if i == nil {
		log.Printf("renderTag NIL! %#v", l)
		return nil
	}
	var class string
	switch i.State {
	case TagInactive:
		class = "is-white"
	case TagUnselected:
		class = "is-light"
	case TagSelected:
		class = "is-info"
	}

	return app.Div().Class("tag m-1").Class(class).Body(
		app.Span().Class("icon-text").Body(
			app.If(i.Icon != nil, app.Span().Class("icon").Body(i.Icon)),
			app.Span().Body(
				app.Text(i.Text),
				app.If(l.cantCount, app.Text(fmt.Sprintf("\u00a0(%d)", i.Count))),
			),
		).OnClick(l.click(i), i.Code),
	)
}

func (l *TagList) click(i *TagInfo) func(ctx app.Context, e app.Event) {
	return func(ctx app.Context, e app.Event) {
		l.Toggle(i.Code)
		l.tagClicked(i)
	}
}

func (l *TagList) tagClicked(i *TagInfo) {
	if l.allTag != nil {
		if i.Code != allCode {
			allSelected := true
			for _, i := range l.tags {
				if i.State == TagUnselected {
					allSelected = false
					continue
				}
			}
			if allSelected {
				l.allTag.State = TagSelected
			} else {
				l.allTag.State = TagUnselected
			}
		} else {
			log.Printf("All tags to %d", i.State)
			for k, t := range l.tags {
				t.State = i.State
				l.tags[k] = t
			}
		}
	}
	l.Update()
}
