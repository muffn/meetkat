//go:build e2e

package e2e

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"meetkat/internal/handler"
	"meetkat/internal/i18n"
	"meetkat/internal/middleware"
	"meetkat/internal/poll"
	"meetkat/internal/view"

	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
)

// screenshotDir is the directory where failure screenshots are saved.
// Set via E2E_SCREENSHOT_DIR; defaults to "" (no screenshots).
var screenshotDir string

// allocCtx is a shared Chrome allocator context, started once in TestMain.
// Each test gets a new browser context (tab) from this single Chrome process.
var allocCtx context.Context

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	headless := os.Getenv("E2E_HEADLESS") != "false"
	screenshotDir = os.Getenv("E2E_SCREENSHOT_DIR")

	if screenshotDir != "" {
		if err := os.MkdirAll(screenshotDir, 0o755); err != nil {
			panic("create screenshot dir: " + err.Error())
		}
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)

	var cancel context.CancelFunc
	allocCtx, cancel = chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Warm up: start Chrome before any tests run.
	warmCtx, warmCancel := chromedp.NewContext(allocCtx)
	if err := chromedp.Run(warmCtx); err != nil {
		panic("failed to start Chrome: " + err.Error())
	}
	warmCancel()

	os.Exit(m.Run())
}

// testServer bundles a running httptest server with its poll service for seeding data.
type testServer struct {
	URL string
	Svc *poll.Service
	srv *httptest.Server
}

// startTestServer creates a Gin router with in-memory repo, real templates, and static
// file serving, then starts an httptest server. The server is automatically closed
// when the test finishes.
func startTestServer(t *testing.T) *testServer {
	t.Helper()

	tr, err := i18n.New()
	if err != nil {
		t.Fatalf("init i18n: %v", err)
	}

	svc := poll.NewService(poll.NewMemoryRepository())
	tmpls := view.LoadTemplates("..")
	ph := handler.NewPollHandler(svc, tmpls)
	hh := handler.NewHomeHandler(tmpls)

	r := gin.New()
	r.Static("/static", "../web/static")
	r.Use(middleware.LangCookie(tr))

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

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	return &testServer{URL: srv.URL, Svc: svc, srv: srv}
}

// seedPoll creates a poll via the service and returns it.
func seedPoll(t *testing.T, svc *poll.Service, title string, options []string) *poll.Poll {
	t.Helper()
	p, err := svc.Create(title, "", options)
	if err != nil {
		t.Fatalf("seedPoll: %v", err)
	}
	return p
}

// newBrowserCtx creates a new browser context (tab) from the shared Chrome process.
// Each test gets an isolated context with a 30-second timeout.
// If E2E_SCREENSHOT_DIR is set, a screenshot is automatically captured on failure.
func newBrowserCtx(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := chromedp.NewContext(allocCtx)
	t.Cleanup(cancel)

	ctx, timeoutCancel := context.WithTimeout(ctx, 30*time.Second)
	t.Cleanup(timeoutCancel)

	// Register screenshot capture before the context cancel cleanups run.
	// Cleanup functions run in LIFO order, so registering this after the
	// cancel cleanups means it runs first (while the browser is still alive).
	screenshotOnFailure(t, ctx)

	return ctx
}

// screenshotOnFailure captures a full-page screenshot when the test has failed.
// It is registered as a cleanup function so it runs after the test body but
// before the browser context is cancelled.
func screenshotOnFailure(t *testing.T, ctx context.Context) {
	t.Helper()
	if screenshotDir == "" {
		return
	}
	t.Cleanup(func() {
		if !t.Failed() {
			return
		}
		var buf []byte
		// Use a short timeout — if the browser is already broken, don't hang.
		captureCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := chromedp.Run(captureCtx, chromedp.FullScreenshot(&buf, 90)); err != nil {
			t.Logf("screenshot capture failed: %v", err)
			return
		}
		name := filepath.Join(screenshotDir, t.Name()+".png")
		if err := os.WriteFile(name, buf, 0o644); err != nil {
			t.Logf("screenshot write failed: %v", err)
			return
		}
		t.Logf("screenshot saved: %s", name)
	})
}

// waitForSelector is a helper that waits for a CSS selector to be present in the DOM.
func waitForSelector(sel string) chromedp.Action {
	return chromedp.WaitVisible(sel, chromedp.ByQuery)
}
