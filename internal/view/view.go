package view

import (
	"html/template"
	"path/filepath"
)

// LoadTemplates parses all page templates against the base layout.
// baseDir is the project root (e.g. "." from main.go, "../.." from internal/handler/ tests).
func LoadTemplates(baseDir string) map[string]*template.Template {
	tmplsDir := filepath.Join(baseDir, "web", "templates")
	base := filepath.Join(tmplsDir, "layouts", "base.html")
	partial := filepath.Join(tmplsDir, "partials", "vote_table.html")

	type page struct {
		name     string
		partials []string
	}
	pages := []page{
		{name: "index.html"},
		{name: "new.html"},
		{name: "poll.html", partials: []string{partial}},
		{name: "admin.html", partials: []string{partial}},
		{name: "404.html"},
	}

	funcs := template.FuncMap{
		"safeHTML": func(s string) template.HTML { return template.HTML(s) }, // #nosec G203 -- trusted server-rendered content only
	}

	tmpls := make(map[string]*template.Template, len(pages))
	for _, pg := range pages {
		files := append([]string{base, filepath.Join(tmplsDir, pg.name)}, pg.partials...)
		tmpls[pg.name] = template.Must(template.New("").Funcs(funcs).ParseFiles(files...))
	}
	return tmpls
}
