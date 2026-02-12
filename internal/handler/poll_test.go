package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"meetkat/internal/i18n"
	"meetkat/internal/middleware"
	"meetkat/internal/poll"
	"meetkat/internal/view"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupLangTestRouter() *gin.Engine {
	tr, err := i18n.New()
	if err != nil {
		panic(err)
	}
	svc := poll.NewService(poll.NewMemoryRepository())
	tmpls := view.LoadTemplates("../..")
	h := NewPollHandler(svc, tmpls)

	r := gin.New()
	r.Use(middleware.LangCookie(tr))
	r.GET("/new", h.ShowNew)
	return r
}

func setupTestRouter() (*gin.Engine, *poll.Service) {
	tr, err := i18n.New()
	if err != nil {
		panic(err)
	}
	svc := poll.NewService(poll.NewMemoryRepository())
	tmpls := view.LoadTemplates("../..")
	h := NewPollHandler(svc, tmpls)

	r := gin.New()
	r.Use(middleware.LangCookie(tr))
	r.GET("/new", h.ShowNew)
	r.POST("/new", h.CreatePoll)
	r.GET("/poll/:id", h.ShowPoll)
	r.POST("/poll/:id/vote", h.SubmitVote)
	r.GET("/poll/:id/admin", h.ShowAdmin)
	r.POST("/poll/:id/admin/remove", h.RemoveVote)
	r.POST("/poll/:id/admin/delete", h.DeletePoll)
	r.POST("/poll/:id/admin/edit", h.UpdateVote)
	return r, svc
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

func TestCreatePollRedirectsToAdmin(t *testing.T) {
	router, svc := setupTestRouter()

	form := url.Values{
		"title":   {"Team dinner"},
		"dates[]": {"2025-06-10", "2025-06-11"},
	}
	w := postForm(router, "/new", form)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
	}

	loc := w.Header().Get("Location")
	if !strings.HasSuffix(loc, "/admin") {
		t.Fatalf("expected redirect to admin page, got %q", loc)
	}

	// Verify the poll was created and the admin ID in the redirect is valid.
	// Extract admin ID from location: /poll/<adminID>/admin
	parts := strings.Split(loc, "/")
	if len(parts) < 4 {
		t.Fatalf("unexpected redirect location format: %q", loc)
	}
	adminID := parts[2]

	p, err := svc.GetByAdminID(adminID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected poll to be found by admin ID")
	}
	if p.Title != "Team dinner" {
		t.Errorf("expected title 'Team dinner', got %q", p.Title)
	}
}

func TestShowAdmin(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPoll(svc, "Admin poll", []string{"Mon", "Tue"})
	_ = svc.AddVote(p.ID, "Alice", map[string]bool{"Mon": true, "Tue": false})

	req := httptest.NewRequest(http.MethodGet, "/poll/"+p.AdminID+"/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Admin poll") {
		t.Error("expected response body to contain poll title")
	}
	if !strings.Contains(body, "Alice") {
		t.Error("expected response body to contain voter name")
	}
	if !strings.Contains(body, "Remove") {
		t.Error("expected response body to contain Remove button")
	}
	if !strings.Contains(body, p.ID) {
		t.Error("expected response body to contain participant link")
	}
}

func TestShowAdminNotFound(t *testing.T) {
	router, _ := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/poll/nonexistent/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestRemoveVoteHandler(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPoll(svc, "Remove test", []string{"Mon"})
	_ = svc.AddVote(p.ID, "Alice", map[string]bool{"Mon": true})
	_ = svc.AddVote(p.ID, "Bob", map[string]bool{"Mon": true})

	form := url.Values{
		"voter_name": {"Alice"},
	}
	w := postForm(router, "/poll/"+p.AdminID+"/admin/remove", form)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
	}
	expectedLoc := "/poll/" + p.AdminID + "/admin"
	if loc := w.Header().Get("Location"); loc != expectedLoc {
		t.Fatalf("expected redirect to %q, got %q", expectedLoc, loc)
	}

	got, _ := svc.Get(p.ID)
	if len(got.Votes) != 1 {
		t.Fatalf("expected 1 vote after removal, got %d", len(got.Votes))
	}
	if got.Votes[0].Name != "Bob" {
		t.Errorf("expected remaining vote to be Bob, got %q", got.Votes[0].Name)
	}
}

func TestDeletePollHandler(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPoll(svc, "Delete me", []string{"Mon"})
	_ = svc.AddVote(p.ID, "Alice", map[string]bool{"Mon": true})

	form := url.Values{}
	w := postForm(router, "/poll/"+p.AdminID+"/admin/delete", form)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/" {
		t.Fatalf("expected redirect to /, got %q", loc)
	}

	got, _ := svc.Get(p.ID)
	if got != nil {
		t.Error("expected poll to be deleted")
	}
}

func TestDeletePollNotFound(t *testing.T) {
	router, _ := setupTestRouter()

	form := url.Values{}
	w := postForm(router, "/poll/nonexistent/admin/delete", form)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUpdateVoteHandler(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPoll(svc, "Edit test", []string{"Mon", "Tue"})
	_ = svc.AddVote(p.ID, "Alice", map[string]bool{"Mon": true, "Tue": false})

	form := url.Values{
		"old_name": {"Alice"},
		"name":     {"Alicia"},
		"vote-Mon": {"no"},
		"vote-Tue": {"yes"},
	}
	w := postForm(router, "/poll/"+p.AdminID+"/admin/edit", form)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
	}

	got, _ := svc.Get(p.ID)
	if len(got.Votes) != 1 {
		t.Fatalf("expected 1 vote, got %d", len(got.Votes))
	}
	v := got.Votes[0]
	if v.Name != "Alicia" {
		t.Errorf("expected name Alicia, got %q", v.Name)
	}
	if v.Responses["Mon"] {
		t.Error("expected Mon to be false")
	}
	if !v.Responses["Tue"] {
		t.Error("expected Tue to be true")
	}
}

func TestUpdateVoteEmptyName(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPoll(svc, "Edit empty", []string{"Mon"})
	_ = svc.AddVote(p.ID, "Alice", map[string]bool{"Mon": true})

	form := url.Values{
		"old_name": {"Alice"},
		"name":     {""},
		"vote-Mon": {"no"},
	}
	w := postForm(router, "/poll/"+p.AdminID+"/admin/edit", form)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect, got %d", w.Code)
	}

	got, _ := svc.Get(p.ID)
	if len(got.Votes) != 1 {
		t.Fatalf("expected 1 vote unchanged, got %d", len(got.Votes))
	}
	if got.Votes[0].Name != "Alice" {
		t.Errorf("expected vote name unchanged as Alice, got %q", got.Votes[0].Name)
	}
}

func TestLangQueryParamSetsCookie(t *testing.T) {
	router := setupLangTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/new?lang=de", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Response should contain German text.
	body := w.Body.String()
	if !strings.Contains(body, "Zurück") {
		t.Error("expected German nav text 'Zurück' in response body")
	}

	// Set-Cookie header should contain meetkat_lang=de.
	cookies := w.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "meetkat_lang" && c.Value == "de" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Set-Cookie meetkat_lang=de in response")
	}
}

func TestLangCookiePersistsWithoutQueryParam(t *testing.T) {
	router := setupLangTestRouter()

	// Request without ?lang= but with the cookie set.
	req := httptest.NewRequest(http.MethodGet, "/new", nil)
	req.AddCookie(&http.Cookie{Name: "meetkat_lang", Value: "de"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Zurück") {
		t.Error("expected German nav text 'Zurück' when cookie is set")
	}
	if strings.Contains(body, ">Back<") {
		t.Error("did not expect English nav text when cookie is de")
	}
}

func TestLangQueryParamOverridesCookie(t *testing.T) {
	router := setupLangTestRouter()

	// Cookie says German, but query param says English.
	req := httptest.NewRequest(http.MethodGet, "/new?lang=en", nil)
	req.AddCookie(&http.Cookie{Name: "meetkat_lang", Value: "de"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	if strings.Contains(body, "Zurück") {
		t.Error("expected English response when ?lang=en overrides cookie")
	}

	// Cookie should be updated to en.
	cookies := w.Result().Cookies()
	var found bool
	for _, c := range cookies {
		if c.Name == "meetkat_lang" && c.Value == "en" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Set-Cookie meetkat_lang=en to override previous cookie")
	}
}
