package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand/v2"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Vote struct {
	Name      string
	Responses map[string]bool // key = date option string, value = available
}

type Poll struct {
	ID        string
	Title     string
	Options   []string
	Votes     []Vote
	CreatedAt time.Time
}

var (
	polls   = make(map[string]*Poll)
	pollsMu sync.Mutex
)

const idChars = "abcdefghijklmnopqrstuvwxyz0123456789"

func generateID() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = idChars[rand.IntN(len(idChars))]
	}
	return string(b)
}

func loadTemplates() map[string]*template.Template {
	base := "templates/layouts/base.html"
	pages := map[string]string{
		"index.html": "templates/index.html",
		"new.html":   "templates/new.html",
		"poll.html":  "templates/poll.html",
	}

	tmpls := make(map[string]*template.Template, len(pages))
	for name, path := range pages {
		tmpls[name] = template.Must(template.ParseFiles(base, path))
	}
	return tmpls
}

func renderHTML(c *gin.Context, tmpls map[string]*template.Template, code int, name string, data any) {
	tmpl, ok := tmpls[name]
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

func main() {
	r := gin.Default()
	tmpls := loadTemplates()

	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		renderHTML(c, tmpls, http.StatusOK, "index.html", gin.H{
			"title": "meetkat",
		})
	})

	r.GET("/new", func(c *gin.Context) {
		renderHTML(c, tmpls, http.StatusOK, "new.html", gin.H{
			"title": "Create a Poll – meetkat",
		})
	})

	r.POST("/new", func(c *gin.Context) {
		title := strings.TrimSpace(c.PostForm("title"))
		dates := c.PostFormArray("dates[]")

		// Filter out empty date values.
		var options []string
		for _, d := range dates {
			d = strings.TrimSpace(d)
			if d != "" {
				options = append(options, d)
			}
		}

		// Validate input.
		var errors []string
		if title == "" {
			errors = append(errors, "Please enter a poll title.")
		}
		if len(options) == 0 {
			errors = append(errors, "Please add at least one date option.")
		}

		if len(errors) > 0 {
			renderHTML(c, tmpls, http.StatusUnprocessableEntity, "new.html", gin.H{
				"title":     "Create a Poll – meetkat",
				"errors":    errors,
				"formTitle": title,
				"formDates": dates,
			})
			return
		}

		id := generateID()
		poll := &Poll{
			ID:        id,
			Title:     title,
			Options:   options,
			CreatedAt: time.Now(),
		}

		pollsMu.Lock()
		polls[id] = poll
		pollsMu.Unlock()

		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/poll/%s", id))
	})

	r.GET("/poll/:id", func(c *gin.Context) {
		id := c.Param("id")

		pollsMu.Lock()
		poll, ok := polls[id]
		pollsMu.Unlock()

		if !ok {
			c.String(http.StatusNotFound, "Poll not found")
			return
		}

		// Compute per-option vote totals.
		totals := make(map[string]int, len(poll.Options))
		for _, opt := range poll.Options {
			for _, v := range poll.Votes {
				if v.Responses[opt] {
					totals[opt]++
				}
			}
		}

		renderHTML(c, tmpls, http.StatusOK, "poll.html", gin.H{
			"title":  poll.Title + " – meetkat",
			"poll":   poll,
			"totals": totals,
			"url":    fmt.Sprintf("%s/poll/%s", c.Request.Host, poll.ID),
		})
	})

	r.POST("/poll/:id/vote", func(c *gin.Context) {
		id := c.Param("id")

		pollsMu.Lock()
		poll, ok := polls[id]
		pollsMu.Unlock()

		if !ok {
			c.String(http.StatusNotFound, "Poll not found")
			return
		}

		name := strings.TrimSpace(c.PostForm("name"))
		if name == "" {
			// Re-render with validation error.
			totals := make(map[string]int, len(poll.Options))
			for _, opt := range poll.Options {
				for _, v := range poll.Votes {
					if v.Responses[opt] {
						totals[opt]++
					}
				}
			}
			renderHTML(c, tmpls, http.StatusUnprocessableEntity, "poll.html", gin.H{
				"title":     poll.Title + " – meetkat",
				"poll":      poll,
				"totals":    totals,
				"url":       fmt.Sprintf("%s/poll/%s", c.Request.Host, poll.ID),
				"voteError": "Please enter your name.",
			})
			return
		}

		responses := make(map[string]bool, len(poll.Options))
		for _, opt := range poll.Options {
			responses[opt] = c.PostForm("vote-"+opt) == "yes"
		}

		pollsMu.Lock()
		poll.Votes = append(poll.Votes, Vote{Name: name, Responses: responses})
		pollsMu.Unlock()

		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/poll/%s", id))
	})

	log.Fatal(r.Run(":8080"))
}
