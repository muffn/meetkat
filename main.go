package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
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

	if err := os.Chmod(cfg.DBPath, 0o600); err != nil {
		slog.Warn("could not set DB file permissions", "err", err)
	}

	translator, err := i18n.New()
	if err != nil {
		log.Fatalf("init i18n: %v", err)
	}

	repo := sqlite.NewPollRepository(db)
	svc := poll.NewService(repo)
	tmpls := view.LoadTemplates(".")
	ph := handler.NewPollHandler(svc, tmpls)
	hh := handler.NewHomeHandler(tmpls)

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	createLimiter := middleware.NewRateLimiter(10, 10)
	voteLimiter := middleware.NewRateLimiter(30, 30)

	r := gin.Default()
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CSRF())
	r.Use(middleware.LangCookie(translator))
	r.Static("/static", "./web/static")
	r.GET("/sw.js", func(c *gin.Context) {
		c.File("./web/static/js/sw.js")
	})

	r.GET("/", hh.ShowHome)
	r.GET("/new", ph.ShowNew)
	r.POST("/new", createLimiter.Middleware(), ph.CreatePoll)
	r.GET("/poll/:id", ph.ShowPoll)
	r.POST("/poll/:id/vote", voteLimiter.Middleware(), ph.SubmitVote)
	r.GET("/poll/:id/vote", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/poll/"+c.Param("id"))
	})
	r.GET("/poll/:id/admin", ph.ShowAdmin)
	r.POST("/poll/:id/admin/vote", voteLimiter.Middleware(), ph.SubmitAdminVote)
	r.GET("/poll/:id/admin/vote", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/poll/"+c.Param("id")+"/admin")
	})
	r.POST("/poll/:id/admin/remove", voteLimiter.Middleware(), ph.RemoveVote)
	r.POST("/poll/:id/admin/delete", voteLimiter.Middleware(), ph.DeletePoll)
	r.POST("/poll/:id/admin/edit", voteLimiter.Middleware(), ph.UpdateVote)

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
