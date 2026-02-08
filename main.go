package main

import (
	"html/template"
	"log"
	"net/http"

	"meetkat/internal/handler"
	"meetkat/internal/poll"

	"github.com/gin-gonic/gin"
)

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

func main() {
	svc := poll.NewService()
	tmpls := loadTemplates()
	h := handler.NewPollHandler(svc, tmpls)

	r := gin.Default()
	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		tmpl, ok := tmpls["index.html"]
		if !ok {
			c.String(http.StatusInternalServerError, "template not found")
			return
		}
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(c.Writer, "index.html", gin.H{
			"title": "meetkat",
		}); err != nil {
			log.Printf("template render error: %v", err)
		}
	})

	r.GET("/new", h.ShowNew)
	r.POST("/new", h.CreatePoll)
	r.GET("/poll/:id", h.ShowPoll)
	r.POST("/poll/:id/vote", h.SubmitVote)

	log.Fatal(r.Run(":8080"))
}
