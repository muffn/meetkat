package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

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

	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o750); err != nil {
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
	r.GET("/sw.js", func(c *gin.Context) {
		c.File("./web/static/js/sw.js")
	})
	r.Use(middleware.LangCookie(translator))

	r.GET("/", hh.ShowHome)
	r.GET("/new", ph.ShowNew)
	r.POST("/new", ph.CreatePoll)
	r.GET("/poll/:id", ph.ShowPoll)
	r.POST("/poll/:id/vote", ph.SubmitVote)
	r.GET("/poll/:id/vote", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/poll/"+c.Param("id"))
	})
	r.GET("/poll/:id/admin", ph.ShowAdmin)
	r.POST("/poll/:id/admin/vote", ph.SubmitAdminVote)
	r.GET("/poll/:id/admin/vote", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/poll/"+c.Param("id")+"/admin")
	})
	r.POST("/poll/:id/admin/remove", ph.RemoveVote)
	r.POST("/poll/:id/admin/delete", ph.DeletePoll)
	r.POST("/poll/:id/admin/edit", ph.UpdateVote)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
}
