package bulma

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// OnInputFn is called when OnInput event is thrown by the element
type OnInputFn func(ctx app.Context, v string)

type TextField struct {
	app.Compo
	Label       string
	Placeholder string
	Value       string
	helpClass   string
	help        string
	OnInput     OnInputFn
}

func NewTextField(value string, label string, placeholder string) *TextField {
	return &TextField{
		Label:       label,
		Placeholder: placeholder,
		Value:       value,
	}
}

// WhitOnInput call fn whenever an Input event is thrown.
// fn retrun set Help part, and error set help style
func (t *TextField) WithOnInput(fn OnInputFn) *TextField {
	t.OnInput = fn
	return t
}
func (t *TextField) OnMount(ctx app.Context) {
	if t.OnInput != nil {
		t.OnInput(ctx, t.Value)
	}
}

func (t *TextField) SetHelp(text, class string) {
	t.help = text
	t.helpClass = class
}

func (t *TextField) Render() app.UI {
	return app.Div().
		Class("field").
		Body(
			app.Label().
				Class("label").
				Text(t.Label),
			app.Div().
				Class("control").
				Body(
					app.Input().
						Class("input").
						Type("text").
						Placeholder(t.Placeholder).
						Value(t.Value).
						OnInput(t.onInput),
				),
			app.P().Class(t.helpClass).Text(t.help),
		)
}

func (t *TextField) onInput(ctx app.Context, e app.Event) {
	t.Value = ctx.JSSrc().Get("value").String()
	if t.OnInput != nil {
		t.OnInput(ctx, t.Value)
	}
}

type option struct {
	value    string
	text     string
	selected bool
}
type SelectField struct {
	app.Compo
	label     string
	value     string
	values    []option
	OnInput   OnInputFn
	help      string
	helpClass string
}

func NewSelectField(value string, label string) *SelectField {
	s := SelectField{
		label: label,
		value: value,
	}
	return &s
}

func (s *SelectField) WithOption(value string, text string, selected bool) *SelectField {
	s.values = append(s.values, option{value, text, selected})
	return s
}

// WhitOnInput call fn whenever an Input event is thrown.
// fn retrun is just ignored
func (s *SelectField) WhitOnInput(fn OnInputFn) *SelectField {
	s.OnInput = fn
	return s
}

func (s *SelectField) OnMount(ctx app.Context) {
	if s.OnInput != nil {
		for _, o := range s.values {
			if o.selected {
				s.OnInput(ctx, o.value)
			}
		}
	}
}
func (s *SelectField) SetHelp(text, class string) {
	s.help = text
	s.helpClass = class
}

func (s *SelectField) Render() app.UI {
	return app.Div().Class("field").Body(
		app.Label().Class("label").Text(s.label),
		app.Div().Class("control").Body(
			app.Div().Class("select").Body(
				app.Select().OnInput(s.onInput, s.label).Body(
					app.Range(s.values).Slice(func(i int) app.UI {
						return app.Option().Value(s.values[i].value).Selected(s.values[i].selected).Text(s.values[i].text)
					}),
				),
			),
		),
		app.If(len(s.help) > 0, app.P().Class("help").Class(s.helpClass).Text(s.help)),
	)
}

func (s *SelectField) onInput(ctx app.Context, e app.Event) {
	s.value = ctx.JSSrc().Get("value").String()
	if s.OnInput != nil {
		s.OnInput(ctx, s.value)
	}
}

type OptionField struct {
	app.Compo
	value    string
	text     string
	selected bool
}

func NewOptionField(value string, text string, selected bool) *OptionField {
	return &OptionField{
		value:    value,
		text:     text,
		selected: selected,
	}
}

func (o *OptionField) Render() app.UI {
	return app.Option().Value(o.value).Selected(o.selected).Text(o.text)
}

type RadioFields struct {
	app.Compo
	label     string
	value     string
	values    []option
	OnInput   OnInputFn
	help      string
	helpClass string
}

func NewRadioFields(value string, label string) *RadioFields {
	s := RadioFields{
		label: label,
		value: value,
	}
	return &s
}

func (s *RadioFields) WithOption(value string, text string, selected bool) *RadioFields {
	s.values = append(s.values, option{value, text, selected})
	return s
}

// WhitOnInput call fn whenever an Input event is thrown.
// fn retrun is just ignored
func (s *RadioFields) WhitOnInput(fn OnInputFn) *RadioFields {
	s.OnInput = fn
	return s
}

func (s *RadioFields) OnMount(ctx app.Context) {
	if s.OnInput != nil {
		for _, o := range s.values {
			if o.selected {
				s.OnInput(ctx, o.value)
			}
		}
	}
}
func (s *RadioFields) SetHelp(text, class string) {
	s.help = text
	s.helpClass = class
}

func (s *RadioFields) Render() app.UI {
	return app.Div().Class("field").Body(
		app.Label().Class("label").Text(s.label),
		app.Div().Class("field").Body(
			app.Div().Class("control").Body(
				app.Range(s.values).Slice(func(i int) app.UI {
					return app.Label().Class("radio").Body(
						app.Input().
							Type("radio").
							Value(s.values[i].value).
							Checked(s.values[i].selected).
							OnInput(s.onInput(s.values[i])),
						app.Text(s.values[i].text),
					)
				}),
			),
		),
		app.If(len(s.help) > 0, app.P().Class("help").Class(s.helpClass).Text(s.help)),
	)
}

func (s *RadioFields) onInput(o option) func(ctx app.Context, e app.Event) {
	return func(ctx app.Context, e app.Event) {
		s.value = o.value
		for i := range s.values {
			s.values[i].selected = s.values[i].value == s.value
		}
		log.Printf("value :%s", s.value)
		if s.OnInput != nil {
			s.OnInput(ctx, s.value)
		}
	}
}

func stringIf(b bool, thenStr, elseStr string) string {
	if b {
		return thenStr
	}
	return elseStr
}

type OnInputTimeFn func(ctx app.Context, v time.Time)
type TimeField struct {
	app.Compo
	Label       string
	fieldType   string
	Placeholder string
	Value       string
	helpClass   string
	help        string
	OnInput     OnInputTimeFn
	list        []time.Time
}

// NewTimeField create an time input field of type fieldType.
// Accepted type :
// 	time
// not yet supported :
//   date
//   date-time-local
//   week
//   month

func NewTimeField(value time.Time, fieldType string, label string, placeholder string) *TimeField {
	t := TimeField{
		Label:       label,
		Placeholder: placeholder,
		fieldType:   fieldType,
	}
	t.Value = t.timeToHTMLString(value)
	return &t
}

func (t *TimeField) List(l []time.Time) *TimeField {
	t.list = l
	return t
}

func (t *TimeField) timeToHTMLString(v time.Time) string {
	switch t.fieldType {
	case "time":
		return v.Format("15:04")
	default:
		return v.String()
	}
}

func (t *TimeField) htmlToTime(s string) (tv time.Time, err error) {
	switch t.fieldType {
	case "time":
		return time.Parse("15:04", s)
	default:
		return time.Parse(time.RFC3339, s)
	}
}

// WhitOnInput call fn whenever an Input event is thrown.
// fn retrun set Help part, and error set help style
func (t *TimeField) WithOnInput(fn OnInputTimeFn) *TimeField {
	t.OnInput = fn
	return t
}
func (t *TimeField) OnMount(ctx app.Context) {
	if t.OnInput != nil {
		tv, _ := t.htmlToTime(t.Value)
		t.OnInput(ctx, tv)
	}
}

func (t *TimeField) SetHelp(text, class string) {
	t.help = text
	t.helpClass = class
}

func (t *TimeField) Render() app.UI {
	var id string
	if t.list != nil {
		id = uuid.NewString()
	}

	return app.Div().
		Class("field").
		Body(
			app.If(t.list != nil,
				app.DataList().ID(id).Body(
					app.Range(t.list).Slice(func(i int) app.UI {
						return app.Option().Value(t.timeToHTMLString(t.list[i]))
					}),
				),
			),
			app.Label().
				Class("label").
				Text(t.Label),
			app.Div().
				Class("control").
				Body(
					app.Input().
						Class("input").
						Type(t.fieldType).
						Placeholder(t.Placeholder).
						Value(t.Value).
						OnInput(t.onInput).
						List(id),
				),
			app.P().Class(t.helpClass).Text(t.help),
		)
}

func (t *TimeField) onInput(ctx app.Context, e app.Event) {
	t.Value = ctx.JSSrc().Get("value").String()
	if t.OnInput != nil {
		tv, _ := t.htmlToTime(t.Value)
		t.OnInput(ctx, tv)
	}
}
