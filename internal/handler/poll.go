package handler

import (
	"fmt"
	"html/template"
	"log/slog"
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

// respondAfterMutation re-fetches the poll via loadPoll and renders the
// vote_table fragment for AJAX requests, or redirects for form POSTs.
func (h *PollHandler) respondAfterMutation(
	c *gin.Context,
	loadPoll func() (*poll.Poll, error),
	isAdmin bool,
	pageName, redirectURL string,
) {
	if isAJAX(c) {
		p, _ := loadPoll()
		h.renderVoteTable(c, p, isAdmin, pageName)
		return
	}
	c.Redirect(http.StatusSeeOther, redirectURL)
}

// renderVoteTable renders only the vote_table fragment for AJAX responses.
func (h *PollHandler) renderVoteTable(c *gin.Context, p *poll.Poll, isAdmin bool, pageName string) {
	totals := poll.Totals(p)
	renderFragment(h.tmpls, c, pageName, "vote_table", gin.H{
		"poll":         p,
		"totals":       totals,
		"winners":      view.WinningOptions(totals),
		"isAdmin":      isAdmin,
		"answerMode":   p.AnswerMode,
		"headerGroups": view.BuildDateHeaders(p.Options, LocalizerFromCtx(c).T),
	})
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
			"formAnswerMode":  c.PostForm("answer_mode"),
		})
		return
	}

	sort.Strings(options)

	answerMode := c.PostForm("answer_mode")

	p, err := h.svc.Create(title, description, answerMode, options)
	if err != nil {
		slog.Error("create poll error", "err", err)
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

// mustLoadPoll fetches a poll by its public or admin ID. On error or not-found
// it writes the appropriate response and returns (nil, false), so the caller
// can simply do:
//
//	p, ok := h.mustLoadPoll(c, id, false)
//	if !ok { return }
func (h *PollHandler) mustLoadPoll(c *gin.Context, id string, isAdmin bool) (*poll.Poll, bool) {
	var p *poll.Poll
	var err error
	if isAdmin {
		p, err = h.svc.GetByAdminID(id)
	} else {
		p, err = h.svc.Get(id)
	}
	if err != nil {
		slog.Error("load poll error", "err", err)
		loc := LocalizerFromCtx(c)
		c.String(http.StatusInternalServerError, loc.T("error.generic"))
		return nil, false
	}
	if p == nil {
		h.renderNotFound(c)
		return nil, false
	}
	return p, true
}

func (h *PollHandler) ShowPoll(c *gin.Context) {
	loc := LocalizerFromCtx(c)
	id := c.Param("id")

	p, ok := h.mustLoadPoll(c, id, false)
	if !ok {
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
		"answerMode":   p.AnswerMode,
		"headerGroups": view.BuildDateHeaders(p.Options, loc.T),
	})
}

func (h *PollHandler) SubmitVote(c *gin.Context) {
	loc := LocalizerFromCtx(c)
	id := c.Param("id")

	p, ok := h.mustLoadPoll(c, id, false)
	if !ok {
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		respondError(c, http.StatusBadRequest, "name required", fmt.Sprintf("/poll/%s", id))
		return
	}

	responses := parseVoteResponses(p.Options, c)

	if err := h.svc.AddVote(id, name, responses); err != nil {
		slog.Error("add vote error", "err", err)
		c.String(http.StatusInternalServerError, loc.T("error.generic"))
		return
	}

	h.respondAfterMutation(c, func() (*poll.Poll, error) { return h.svc.Get(id) }, false, "poll.html", fmt.Sprintf("/poll/%s", id))
}

func (h *PollHandler) ShowAdmin(c *gin.Context) {
	loc := LocalizerFromCtx(c)
	adminID := c.Param("id")

	p, ok := h.mustLoadPoll(c, adminID, true)
	if !ok {
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
		"answerMode":   p.AnswerMode,
		"headerGroups": view.BuildDateHeaders(p.Options, loc.T),
	})
}

func (h *PollHandler) SubmitAdminVote(c *gin.Context) {
	loc := LocalizerFromCtx(c)
	adminID := c.Param("id")

	p, ok := h.mustLoadPoll(c, adminID, true)
	if !ok {
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		respondError(c, http.StatusBadRequest, "name required", fmt.Sprintf("/poll/%s/admin", adminID))
		return
	}

	responses := parseVoteResponses(p.Options, c)

	if err := h.svc.AddVote(p.ID, name, responses); err != nil {
		slog.Error("add vote error", "err", err)
		c.String(http.StatusInternalServerError, loc.T("error.generic"))
		return
	}

	h.respondAfterMutation(c, func() (*poll.Poll, error) { return h.svc.GetByAdminID(adminID) }, true, "admin.html", fmt.Sprintf("/poll/%s/admin", adminID))
}

func (h *PollHandler) RemoveVote(c *gin.Context) {
	adminID := c.Param("id")

	p, ok := h.mustLoadPoll(c, adminID, true)
	if !ok {
		return
	}

	voterName := strings.TrimSpace(c.PostForm("voter_name"))
	if voterName == "" {
		respondError(c, http.StatusBadRequest, "voter_name required", fmt.Sprintf("/poll/%s/admin", adminID))
		return
	}

	if err := h.svc.RemoveVote(p.ID, voterName); err != nil {
		slog.Error("remove vote error", "err", err)
	}

	h.respondAfterMutation(c, func() (*poll.Poll, error) { return h.svc.GetByAdminID(adminID) }, true, "admin.html", fmt.Sprintf("/poll/%s/admin", adminID))
}

func (h *PollHandler) DeletePoll(c *gin.Context) {
	loc := LocalizerFromCtx(c)
	adminID := c.Param("id")

	p, ok := h.mustLoadPoll(c, adminID, true)
	if !ok {
		return
	}

	if err := h.svc.Delete(p.ID); err != nil {
		slog.Error("delete poll error", "err", err)
		c.String(http.StatusInternalServerError, loc.T("error.generic"))
		return
	}

	c.Redirect(http.StatusSeeOther, "/")
}

func (h *PollHandler) UpdateVote(c *gin.Context) {
	adminID := c.Param("id")

	p, ok := h.mustLoadPoll(c, adminID, true)
	if !ok {
		return
	}

	oldName := strings.TrimSpace(c.PostForm("old_name"))
	newName := strings.TrimSpace(c.PostForm("name"))

	if newName == "" {
		respondError(c, http.StatusBadRequest, "name required", fmt.Sprintf("/poll/%s/admin", adminID))
		return
	}

	responses := parseVoteResponses(p.Options, c)

	if err := h.svc.UpdateVote(p.ID, oldName, newName, responses); err != nil {
		slog.Error("update vote error", "err", err)
	}

	h.respondAfterMutation(c, func() (*poll.Poll, error) { return h.svc.GetByAdminID(adminID) }, true, "admin.html", fmt.Sprintf("/poll/%s/admin", adminID))
}

// parseVoteResponses reads vote-<option> form values and returns a response map.
// Valid values are "yes", "maybe", "no"; anything else defaults to "no".
func parseVoteResponses(options []string, c *gin.Context) map[string]string {
	responses := make(map[string]string, len(options))
	for _, opt := range options {
		val := c.PostForm("vote-" + opt)
		switch val {
		case "yes", "maybe":
			responses[opt] = val
		default:
			responses[opt] = "no"
		}
	}
	return responses
}
