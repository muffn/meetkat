package handler

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"meetkat/internal/poll"

	"github.com/gin-gonic/gin"
)

type PollHandler struct {
	svc   *poll.Service
	tmpls map[string]*template.Template
}

func NewPollHandler(svc *poll.Service, tmpls map[string]*template.Template) *PollHandler {
	return &PollHandler{svc: svc, tmpls: tmpls}
}

func (h *PollHandler) renderHTML(c *gin.Context, code int, name string, data any) {
	tmpl, ok := h.tmpls[name]
	if !ok {
		c.String(http.StatusInternalServerError, "template %q not found", name)
		return
	}
	c.Status(code)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, name, data); err != nil {
		log.Printf("template render error: %v", err)
	}
}

func (h *PollHandler) ShowNew(c *gin.Context) {
	h.renderHTML(c, http.StatusOK, "new.html", gin.H{
		"title": "Create a Poll – meetkat",
	})
}

func (h *PollHandler) CreatePoll(c *gin.Context) {
	title := strings.TrimSpace(c.PostForm("title"))
	description := strings.TrimSpace(c.PostForm("description"))
	dates := c.PostFormArray("dates[]")

	var options []string
	for _, d := range dates {
		d = strings.TrimSpace(d)
		if d != "" {
			options = append(options, d)
		}
	}

	var errors []string
	if title == "" {
		errors = append(errors, "Please enter a poll title.")
	}
	if len(options) == 0 {
		errors = append(errors, "Please add at least one date option.")
	}

	if len(errors) > 0 {
		h.renderHTML(c, http.StatusUnprocessableEntity, "new.html", gin.H{
			"title":           "Create a Poll – meetkat",
			"errors":          errors,
			"formTitle":       title,
			"formDescription": description,
			"formDates":       dates,
		})
		return
	}

	p := h.svc.Create(title, description, options)
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/poll/%s", p.ID))
}

func (h *PollHandler) renderNotFound(c *gin.Context) {
	h.renderHTML(c, http.StatusNotFound, "404.html", gin.H{
		"title": "Poll Not Found – meetkat",
	})
}

func (h *PollHandler) ShowPoll(c *gin.Context) {
	id := c.Param("id")

	p, ok := h.svc.Get(id)
	if !ok {
		h.renderNotFound(c)
		return
	}

	totals := poll.Totals(p)
	h.renderHTML(c, http.StatusOK, "poll.html", gin.H{
		"title":  p.Title + " – meetkat",
		"poll":   p,
		"totals": totals,
		"url":    fmt.Sprintf("%s/poll/%s", c.Request.Host, p.ID),
	})
}

func (h *PollHandler) SubmitVote(c *gin.Context) {
	id := c.Param("id")

	p, ok := h.svc.Get(id)
	if !ok {
		h.renderNotFound(c)
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		totals := poll.Totals(p)
		h.renderHTML(c, http.StatusUnprocessableEntity, "poll.html", gin.H{
			"title":     p.Title + " – meetkat",
			"poll":      p,
			"totals":    totals,
			"url":       fmt.Sprintf("%s/poll/%s", c.Request.Host, p.ID),
			"voteError": "Please enter your name.",
		})
		return
	}

	responses := make(map[string]bool, len(p.Options))
	for _, opt := range p.Options {
		responses[opt] = c.PostForm("vote-"+opt) == "yes"
	}

	// AddVote won't fail here: name is non-empty and poll exists.
	_ = h.svc.AddVote(id, name, responses)

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/poll/%s", id))
}
