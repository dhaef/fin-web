package tui

import (
	"bytes"
	"fmt"
	"text/template"

	tea "charm.land/bubbletea/v2"
)

type Route struct {
	Key   string
	Value tea.Model
}

func NewRoute(key string, value tea.Model) Route {
	return Route{key, value}
}

type AppParams struct {
	IsNavigating bool
	CurrentRoute int
	Routes       []Route
	View         string
}

type App struct {
	CurrModel    tea.Model
	Routes       []Route
	currentRoute int
	isNavigating bool
}

func NewApp(routes ...Route) App {
	app := App{Routes: routes}
	if len(routes) > 0 {
		app.CurrModel = routes[0].Value
	}
	return app
}

func (a App) Init() tea.Cmd {
	return tea.ClearScreen
}

func (a App) View() tea.View {
	tmpl := `{{.View}}

{{ if .IsNavigating -}}
	{{ $currentRoute := .CurrentRoute }}
	{{ range $i, $v := .Routes -}}
		{{ if eq $i $currentRoute }}→ {{ else }}  {{ end }}{{ $v.Key }}
	{{ end -}}
{{ end }}
      Exit: ctrl+c
  Navigate: ctrl+n
`
	params := AppParams{IsNavigating: a.isNavigating, CurrentRoute: a.currentRoute, Routes: a.Routes}

	if a.CurrModel != nil {
		params.View = a.CurrModel.View().Content
	}

	t, _ := template.New("main").Parse(tmpl)
	var buf bytes.Buffer
	err := t.Execute(&buf, params)
	if err != nil {
		fmt.Println(err.Error())
	}

	v := tea.NewView(buf.String())
	v.AltScreen = true
	return v
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "ctrl+n":
			a.isNavigating = !a.isNavigating
			return a, nil
		case "up", "down", "enter":
			if !a.isNavigating {
				break
			}
			switch msg.String() {
			case "down":
				if a.currentRoute < len(a.Routes)-1 {
					a.currentRoute += 1
				}
			case "up":
				if a.currentRoute > 0 {
					a.currentRoute -= 1
				}
			case "enter":
				a.CurrModel = a.Routes[a.currentRoute].Value
			}
		}
	}

	if a.CurrModel != nil {
		var cmd tea.Cmd
		a.CurrModel, cmd = a.CurrModel.Update(msg)
		return a, cmd
	}
	return a, nil
}
