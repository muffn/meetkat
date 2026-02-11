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

	p, err := h.svc.Create(title, description, options)
	if err != nil {
		log.Printf("create poll error: %v", err)
		c.String(http.StatusInternalServerError, "Something went wrong. Please try again.")
		return
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/poll/%s/admin", p.AdminID))
}

func (h *PollHandler) renderNotFound(c *gin.Context) {
	h.renderHTML(c, http.StatusNotFound, "404.html", gin.H{
		"title": "Poll Not Found – meetkat",
	})
}

func (h *PollHandler) ShowPoll(c *gin.Context) {
	id := c.Param("id")

	p, err := h.svc.Get(id)
	if err != nil {
		log.Printf("get poll error: %v", err)
		c.String(http.StatusInternalServerError, "Something went wrong. Please try again.")
		return
	}
	if p == nil {
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

	p, err := h.svc.Get(id)
	if err != nil {
		log.Printf("get poll error: %v", err)
		c.String(http.StatusInternalServerError, "Something went wrong. Please try again.")
		return
	}
	if p == nil {
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

	if err := h.svc.AddVote(id, name, responses); err != nil {
		log.Printf("add vote error: %v", err)
		c.String(http.StatusInternalServerError, "Something went wrong. Please try again.")
		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/poll/%s", id))
}

func (h *PollHandler) ShowAdmin(c *gin.Context) {
	adminID := c.Param("id")

	p, err := h.svc.GetByAdminID(adminID)
	if err != nil {
		log.Printf("get poll by admin id error: %v", err)
		c.String(http.StatusInternalServerError, "Something went wrong. Please try again.")
		return
	}
	if p == nil {
		h.renderNotFound(c)
		return
	}

	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	totals := poll.Totals(p)
	h.renderHTML(c, http.StatusOK, "admin.html", gin.H{
		"title":    p.Title + " – Admin – meetkat",
		"poll":     p,
		"totals":   totals,
		"pollURL":  fmt.Sprintf("%s/poll/%s", baseURL, p.ID),
		"adminURL": fmt.Sprintf("%s/poll/%s/admin", baseURL, p.AdminID),
	})
}

func (h *PollHandler) RemoveVote(c *gin.Context) {
	adminID := c.Param("id")

	p, err := h.svc.GetByAdminID(adminID)
	if err != nil {
		log.Printf("get poll by admin id error: %v", err)
		c.String(http.StatusInternalServerError, "Something went wrong. Please try again.")
		return
	}
	if p == nil {
		h.renderNotFound(c)
		return
	}

	voterName := strings.TrimSpace(c.PostForm("voter_name"))
	if voterName == "" {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/poll/%s/admin", adminID))
		return
	}

	if err := h.svc.RemoveVote(p.ID, voterName); err != nil {
		log.Printf("remove vote error: %v", err)
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/poll/%s/admin", adminID))
}
