package handler

import (
	"html/template"
	"log"
	"net/http"

	"meetkat/internal/i18n"

	"github.com/gin-gonic/gin"
)

// LocalizerFromCtx retrieves the *i18n.Localizer set by the language middleware.
func LocalizerFromCtx(c *gin.Context) *i18n.Localizer {
	return c.MustGet("localizer").(*i18n.Localizer)
}

// isAJAX returns true when the request was made via fetch() with our custom header.
func isAJAX(c *gin.Context) bool {
	return c.GetHeader("X-Requested-With") == "fetch"
}

// renderFragment renders only the named fragment (e.g. "vote_table") from the
// page template, without the base layout wrapper.  Used to return partial HTML
// for AJAX requests.
func renderFragment(tmpls map[string]*template.Template, c *gin.Context, pageName, fragmentName string, data gin.H) {
	tmpl, ok := tmpls[pageName]
	if !ok {
		c.String(http.StatusInternalServerError, "template %q not found", pageName)
		return
	}
	loc := LocalizerFromCtx(c)
	data["t"] = loc.T
	data["lang"] = loc.Lang()
	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, fragmentName, data); err != nil {
		log.Printf("fragment render error: %v", err)
	}
}

// renderHTML executes the named template with locale data injected.
func renderHTML(tmpls map[string]*template.Template, c *gin.Context, code int, name string, data gin.H) {
	tmpl, ok := tmpls[name]
	if !ok {
		c.String(http.StatusInternalServerError, "template %q not found", name)
		return
	}
	loc := LocalizerFromCtx(c)
	data["t"] = loc.T
	data["lang"] = loc.Lang()
	c.Status(code)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, name, data); err != nil {
		log.Printf("template render error: %v", err)
	}
}
