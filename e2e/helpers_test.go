//go:build e2e

package e2e

import (
	"context"
	"net/http/httptest"
	"os"
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

func init() {
	gin.SetMode(gin.TestMode)
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

// newBrowserCtx creates a Chrome context with a 30-second timeout.
// Set E2E_HEADLESS=false to watch the browser during tests.
// It is automatically cancelled when the test finishes.
func newBrowserCtx(t *testing.T) context.Context {
	t.Helper()

	headless := os.Getenv("E2E_HEADLESS") != "false"

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	t.Cleanup(allocCancel)

	ctx, cancel := chromedp.NewContext(allocCtx)
	t.Cleanup(cancel)

	ctx, timeoutCancel := context.WithTimeout(ctx, 30*time.Second)
	t.Cleanup(timeoutCancel)

	return ctx
}

// waitForSelector is a helper that waits for a CSS selector to be present in the DOM.
func waitForSelector(sel string) chromedp.Action {
	return chromedp.WaitVisible(sel, chromedp.ByQuery)
}
