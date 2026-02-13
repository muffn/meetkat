package handler

import (
	"html/template"
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
	renderHTML(h.tmpls, c, http.StatusOK, "index.html", gin.H{
		"title": "meetkat",
	})
}
