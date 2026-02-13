package main

import (
	"log"
	"os"
	"path/filepath"

	"meetkat/internal/config"
	"meetkat/internal/handler"
	"meetkat/internal/i18n"
	"meetkat/internal/middleware"
	"meetkat/internal/poll"
	"meetkat/internal/sqlite"
	"meetkat/internal/view"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		log.Fatalf("create data directory: %v", err)
	}

	db, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	translator, err := i18n.New()
	if err != nil {
		log.Fatalf("init i18n: %v", err)
	}

	repo := sqlite.NewPollRepository(db)
	svc := poll.NewService(repo)
	tmpls := view.LoadTemplates(".")
	ph := handler.NewPollHandler(svc, tmpls)
	hh := handler.NewHomeHandler(tmpls)

	r := gin.Default()
	r.Static("/static", "./web/static")
	r.Use(middleware.LangCookie(translator))

	r.GET("/", hh.ShowHome)
	r.GET("/new", ph.ShowNew)
	r.POST("/new", ph.CreatePoll)
	r.GET("/poll/:id", ph.ShowPoll)
	r.POST("/poll/:id/vote", ph.SubmitVote)
	r.GET("/poll/:id/admin", ph.ShowAdmin)
	r.POST("/poll/:id/admin/vote", ph.SubmitAdminVote)
	r.POST("/poll/:id/admin/remove", ph.RemoveVote)
	r.POST("/poll/:id/admin/delete", ph.DeletePoll)
	r.POST("/poll/:id/admin/edit", ph.UpdateVote)

	log.Fatal(r.Run(":" + cfg.Port))
}
