package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"meetkat/internal/handler"
	"meetkat/internal/poll"
	"meetkat/internal/sqlite"

	"github.com/gin-gonic/gin"
)

func loadTemplates() map[string]*template.Template {
	base := "templates/layouts/base.html"
	pages := map[string]string{
		"index.html": "templates/index.html",
		"new.html":   "templates/new.html",
		"poll.html":  "templates/poll.html",
		"admin.html": "templates/admin.html",
		"404.html":   "templates/404.html",
	}

	tmpls := make(map[string]*template.Template, len(pages))
	for name, path := range pages {
		tmpls[name] = template.Must(template.ParseFiles(base, path))
	}
	return tmpls
}

func main() {
	dbPath := os.Getenv("MEETKAT_DB_PATH")
	if dbPath == "" {
		dbPath = "data/meetkat.db"
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		log.Fatalf("create data directory: %v", err)
	}

	db, err := sqlite.Open(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	repo := sqlite.NewPollRepository(db)
	svc := poll.NewService(repo)
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
	r.GET("/poll/:id/admin", h.ShowAdmin)
	r.POST("/poll/:id/admin/remove", h.RemoveVote)

	log.Fatal(r.Run(":8080"))
}
