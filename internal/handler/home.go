package handler

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HomeHandler serves the landing page.
type HomeHandler struct {
	tmpls map[string]*template.Template
}

// NewHomeHandler creates a HomeHandler with the given template map.
func NewHomeHandler(tmpls map[string]*template.Template) *HomeHandler {
	return &HomeHandler{tmpls: tmpls}
}

// ShowHome renders the index / hero page.
func (h *HomeHandler) ShowHome(c *gin.Context) {
	tmpl, ok := h.tmpls["index.html"]
	if !ok {
		c.String(http.StatusInternalServerError, "template not found")
		return
	}
	loc := LocalizerFromCtx(c)
	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, "index.html", gin.H{
		"title": "meetkat",
		"t":     loc.T,
		"lang":  loc.Lang(),
	}); err != nil {
		log.Printf("template render error: %v", err)
	}
}
