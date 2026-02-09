package handler

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"meetkat/internal/poll"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestRouter() (*gin.Engine, *poll.Service) {
	svc := poll.NewService(poll.NewMemoryRepository())
	tmpls := loadTestTemplates()
	h := NewPollHandler(svc, tmpls)

	r := gin.New()
	r.GET("/new", h.ShowNew)
	r.POST("/new", h.CreatePoll)
	r.GET("/poll/:id", h.ShowPoll)
	r.POST("/poll/:id/vote", h.SubmitVote)
	return r, svc
}

func loadTestTemplates() map[string]*template.Template {
	base := "../../templates/layouts/base.html"
	pages := map[string]string{
		"index.html": "../../templates/index.html",
		"new.html":   "../../templates/new.html",
		"poll.html":  "../../templates/poll.html",
		"404.html":   "../../templates/404.html",
	}

	tmpls := make(map[string]*template.Template, len(pages))
	for name, path := range pages {
		tmpls[name] = template.Must(template.ParseFiles(base, path))
	}
	return tmpls
}

func postForm(router http.Handler, path string, form url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// seedPoll creates a poll directly via the service for testing.
func seedPoll(svc *poll.Service, title string, options []string) *poll.Poll {
	p, err := svc.Create(title, "", options)
	if err != nil {
		panic(err)
	}
	return p
}

func TestVoteSubmission(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPoll(svc, "Dinner", []string{"2025-06-10", "2025-06-11"})

	form := url.Values{
		"name":            {"Alice"},
		"vote-2025-06-10": {"yes"},
	}
	w := postForm(router, "/poll/"+p.ID+"/vote", form)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect 303, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/poll/"+p.ID {
		t.Fatalf("expected redirect to /poll/%s, got %q", p.ID, loc)
	}

	got, _ := svc.Get(p.ID)
	if len(got.Votes) != 1 {
		t.Fatalf("expected 1 vote, got %d", len(got.Votes))
	}
	v := got.Votes[0]
	if v.Name != "Alice" {
		t.Errorf("expected name Alice, got %q", v.Name)
	}
	if !v.Responses["2025-06-10"] {
		t.Error("expected 2025-06-10 to be true")
	}
	if v.Responses["2025-06-11"] {
		t.Error("expected 2025-06-11 to be false")
	}
}

func TestVoteEmptyName(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPoll(svc, "Lunch", []string{"2025-07-01"})

	form := url.Values{
		"name":            {""},
		"vote-2025-07-01": {"yes"},
	}
	w := postForm(router, "/poll/"+p.ID+"/vote", form)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
	if body := w.Body.String(); !strings.Contains(body, "Please enter your name.") {
		t.Error("expected validation error in response body")
	}

	got, _ := svc.Get(p.ID)
	if len(got.Votes) != 0 {
		t.Error("vote should not have been saved")
	}
}

func TestVoteWhitespaceOnlyName(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPoll(svc, "Meeting", []string{"2025-08-01"})

	form := url.Values{
		"name": {"   "},
	}
	w := postForm(router, "/poll/"+p.ID+"/vote", form)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}

	got, _ := svc.Get(p.ID)
	if len(got.Votes) != 0 {
		t.Error("vote should not have been saved")
	}
}

func TestVoteNonexistentPoll(t *testing.T) {
	router, _ := setupTestRouter()

	form := url.Values{"name": {"Alice"}}
	w := postForm(router, "/poll/doesnotexist/vote", form)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestVoteResponseValues(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPoll(svc, "Offsite", []string{"Mon", "Tue", "Wed"})

	form := url.Values{
		"name":     {"Bob"},
		"vote-Mon": {"yes"},
		"vote-Tue": {"no"},
		// Wed omitted entirely
	}
	w := postForm(router, "/poll/"+p.ID+"/vote", form)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
	}

	got, _ := svc.Get(p.ID)
	v := got.Votes[0]
	tests := []struct {
		option string
		want   bool
	}{
		{"Mon", true},
		{"Tue", false},
		{"Wed", false},
	}
	for _, tt := range tests {
		if got := v.Responses[tt.option]; got != tt.want {
			t.Errorf("option %q: got %v, want %v", tt.option, got, tt.want)
		}
	}
}

func TestMultipleVotes(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPoll(svc, "Sprint planning", []string{"2025-09-01", "2025-09-02"})

	postForm(router, "/poll/"+p.ID+"/vote", url.Values{
		"name":            {"Alice"},
		"vote-2025-09-01": {"yes"},
		"vote-2025-09-02": {"yes"},
	})
	postForm(router, "/poll/"+p.ID+"/vote", url.Values{
		"name":            {"Bob"},
		"vote-2025-09-01": {"yes"},
	})

	got, _ := svc.Get(p.ID)
	if len(got.Votes) != 2 {
		t.Fatalf("expected 2 votes, got %d", len(got.Votes))
	}
	if got.Votes[0].Name != "Alice" {
		t.Errorf("first vote should be Alice, got %q", got.Votes[0].Name)
	}
	if got.Votes[1].Name != "Bob" {
		t.Errorf("second vote should be Bob, got %q", got.Votes[1].Name)
	}
}

func TestPollViewWithVotes(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPoll(svc, "Team dinner", []string{"2025-10-01", "2025-10-02"})

	_ = svc.AddVote(p.ID, "Alice", map[string]bool{"2025-10-01": true, "2025-10-02": false})
	_ = svc.AddVote(p.ID, "Bob", map[string]bool{"2025-10-01": true, "2025-10-02": true})

	req := httptest.NewRequest(http.MethodGet, "/poll/"+p.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	for _, name := range []string{"Alice", "Bob"} {
		if !strings.Contains(body, name) {
			t.Errorf("expected response body to contain %q", name)
		}
	}
}

func TestPollViewNotFound(t *testing.T) {
	router, _ := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/poll/nope", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
