package view

import (
	"html/template"
	"path/filepath"
)

// LoadTemplates parses all page templates against the base layout.
// baseDir is the project root (e.g. "." from main.go, "../.." from internal/handler/ tests).
func LoadTemplates(baseDir string) map[string]*template.Template {
	base := filepath.Join(baseDir, "web", "templates", "layouts", "base.html")
	pages := map[string]string{
		"index.html": filepath.Join(baseDir, "web", "templates", "index.html"),
		"new.html":   filepath.Join(baseDir, "web", "templates", "new.html"),
		"poll.html":  filepath.Join(baseDir, "web", "templates", "poll.html"),
		"admin.html": filepath.Join(baseDir, "web", "templates", "admin.html"),
		"404.html":   filepath.Join(baseDir, "web", "templates", "404.html"),
	}

	funcs := template.FuncMap{
		"safeHTML": func(s string) template.HTML { return template.HTML(s) }, // #nosec G203 -- trusted server-rendered content only
	}

	partial := filepath.Join(baseDir, "web", "templates", "partials", "vote_table.html")

	// Pages that need the vote_table partial
	needsPartial := map[string]bool{
		"poll.html":  true,
		"admin.html": true,
	}

	tmpls := make(map[string]*template.Template, len(pages))
	for name, path := range pages {
		if needsPartial[name] {
			tmpls[name] = template.Must(template.New("").Funcs(funcs).ParseFiles(base, path, partial))
		} else {
			tmpls[name] = template.Must(template.New("").Funcs(funcs).ParseFiles(base, path))
		}
	}
	return tmpls
}
