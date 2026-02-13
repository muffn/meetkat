package handler

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"meetkat/internal/poll"
	"meetkat/internal/view"

	"github.com/gin-gonic/gin"
)

type PollHandler struct {
	svc   *poll.Service
	tmpls map[string]*template.Template
}

func NewPollHandler(svc *poll.Service, tmpls map[string]*template.Template) *PollHandler {
	return &PollHandler{svc: svc, tmpls: tmpls}
}

func (h *PollHandler) ShowNew(c *gin.Context) {
	loc := LocalizerFromCtx(c)
	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	renderHTML(h.tmpls, c, http.StatusOK, "new.html", gin.H{
		"title":     loc.T("new.page_title"),
		"formDates": []string{today, tomorrow},
	})
}

func (h *PollHandler) CreatePoll(c *gin.Context) {
	loc := LocalizerFromCtx(c)
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
		errors = append(errors, loc.T("new.error_no_title"))
	}
	if len(options) == 0 {
		errors = append(errors, loc.T("new.error_no_dates"))
	}

	if len(errors) > 0 {
		renderHTML(h.tmpls, c, http.StatusUnprocessableEntity, "new.html", gin.H{
			"title":           loc.T("new.page_title"),
			"errors":          errors,
			"formTitle":       title,
			"formDescription": description,
			"formDates":       dates,
		})
		return
	}

	sort.Strings(options)

	p, err := h.svc.Create(title, description, options)
	if err != nil {
		log.Printf("create poll error: %v", err)
		c.String(http.StatusInternalServerError, loc.T("error.generic"))
		return
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/poll/%s/admin", p.AdminID))
}

func (h *PollHandler) renderNotFound(c *gin.Context) {
	loc := LocalizerFromCtx(c)
	renderHTML(h.tmpls, c, http.StatusNotFound, "404.html", gin.H{
		"title": loc.T("notfound.page_title"),
	})
}

func (h *PollHandler) ShowPoll(c *gin.Context) {
	loc := LocalizerFromCtx(c)
	id := c.Param("id")

	p, err := h.svc.Get(id)
	if err != nil {
		log.Printf("get poll error: %v", err)
		c.String(http.StatusInternalServerError, loc.T("error.generic"))
		return
	}
	if p == nil {
		h.renderNotFound(c)
		return
	}

	totals := poll.Totals(p)
	renderHTML(h.tmpls, c, http.StatusOK, "poll.html", gin.H{
		"title":        fmt.Sprintf(loc.T("poll.page_title"), p.Title),
		"poll":         p,
		"totals":       totals,
		"winners":      view.WinningOptions(totals),
		"url":          fmt.Sprintf("%s/poll/%s", c.Request.Host, p.ID),
		"isAdmin":      false,
		"headerGroups": view.BuildDateHeaders(p.Options, loc.T),
	})
}

func (h *PollHandler) SubmitVote(c *gin.Context) {
	loc := LocalizerFromCtx(c)
	id := c.Param("id")

	p, err := h.svc.Get(id)
	if err != nil {
		log.Printf("get poll error: %v", err)
		c.String(http.StatusInternalServerError, loc.T("error.generic"))
		return
	}
	if p == nil {
		h.renderNotFound(c)
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/poll/%s", id))
		return
	}

	responses := make(map[string]bool, len(p.Options))
	for _, opt := range p.Options {
		responses[opt] = c.PostForm("vote-"+opt) == "yes"
	}

	if err := h.svc.AddVote(id, name, responses); err != nil {
		log.Printf("add vote error: %v", err)
		c.String(http.StatusInternalServerError, loc.T("error.generic"))
		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/poll/%s", id))
}

func (h *PollHandler) ShowAdmin(c *gin.Context) {
	loc := LocalizerFromCtx(c)
	adminID := c.Param("id")

	p, err := h.svc.GetByAdminID(adminID)
	if err != nil {
		log.Printf("get poll by admin id error: %v", err)
		c.String(http.StatusInternalServerError, loc.T("error.generic"))
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
	renderHTML(h.tmpls, c, http.StatusOK, "admin.html", gin.H{
		"title":        fmt.Sprintf(loc.T("admin.page_title"), p.Title),
		"poll":         p,
		"totals":       totals,
		"winners":      view.WinningOptions(totals),
		"pollURL":      fmt.Sprintf("%s/poll/%s", baseURL, p.ID),
		"adminURL":     fmt.Sprintf("%s/poll/%s/admin", baseURL, p.AdminID),
		"isAdmin":      true,
		"headerGroups": view.BuildDateHeaders(p.Options, loc.T),
	})
}

func (h *PollHandler) SubmitAdminVote(c *gin.Context) {
	loc := LocalizerFromCtx(c)
	adminID := c.Param("id")

	p, err := h.svc.GetByAdminID(adminID)
	if err != nil {
		log.Printf("get poll by admin id error: %v", err)
		c.String(http.StatusInternalServerError, loc.T("error.generic"))
		return
	}
	if p == nil {
		h.renderNotFound(c)
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/poll/%s/admin", adminID))
		return
	}

	responses := make(map[string]bool, len(p.Options))
	for _, opt := range p.Options {
		responses[opt] = c.PostForm("vote-"+opt) == "yes"
	}

	if err := h.svc.AddVote(p.ID, name, responses); err != nil {
		log.Printf("add vote error: %v", err)
		c.String(http.StatusInternalServerError, loc.T("error.generic"))
		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/poll/%s/admin", adminID))
}

func (h *PollHandler) RemoveVote(c *gin.Context) {
	loc := LocalizerFromCtx(c)
	adminID := c.Param("id")

	p, err := h.svc.GetByAdminID(adminID)
	if err != nil {
		log.Printf("get poll by admin id error: %v", err)
		c.String(http.StatusInternalServerError, loc.T("error.generic"))
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

func (h *PollHandler) DeletePoll(c *gin.Context) {
	loc := LocalizerFromCtx(c)
	adminID := c.Param("id")

	p, err := h.svc.GetByAdminID(adminID)
	if err != nil {
		log.Printf("get poll by admin id error: %v", err)
		c.String(http.StatusInternalServerError, loc.T("error.generic"))
		return
	}
	if p == nil {
		h.renderNotFound(c)
		return
	}

	if err := h.svc.Delete(p.ID); err != nil {
		log.Printf("delete poll error: %v", err)
		c.String(http.StatusInternalServerError, loc.T("error.generic"))
		return
	}

	c.Redirect(http.StatusSeeOther, "/")
}

func (h *PollHandler) UpdateVote(c *gin.Context) {
	loc := LocalizerFromCtx(c)
	adminID := c.Param("id")

	p, err := h.svc.GetByAdminID(adminID)
	if err != nil {
		log.Printf("get poll by admin id error: %v", err)
		c.String(http.StatusInternalServerError, loc.T("error.generic"))
		return
	}
	if p == nil {
		h.renderNotFound(c)
		return
	}

	oldName := strings.TrimSpace(c.PostForm("old_name"))
	newName := strings.TrimSpace(c.PostForm("name"))

	if newName == "" {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/poll/%s/admin", adminID))
		return
	}

	responses := make(map[string]bool, len(p.Options))
	for _, opt := range p.Options {
		responses[opt] = c.PostForm("vote-"+opt) == "yes"
	}

	if err := h.svc.UpdateVote(p.ID, oldName, newName, responses); err != nil {
		log.Printf("update vote error: %v", err)
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/poll/%s/admin", adminID))
}
