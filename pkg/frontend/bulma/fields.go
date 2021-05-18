package bulma

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// OnInputFn is called when OnInput event is thrown by the element
type OnInputFn func(v string)

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
		t.OnInput(t.Value)
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
	t.Value = ctx.JSSrc.Get("value").String()
	if t.OnInput != nil {
		t.OnInput(t.Value)
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
				s.OnInput(o.value)
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
	s.value = ctx.JSSrc.Get("value").String()
	if s.OnInput != nil {
		s.OnInput(s.value)
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
