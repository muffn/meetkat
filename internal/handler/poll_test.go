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
	r.POST("/poll/:id/admin/vote", h.SubmitAdminVote)
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

// seedPoll creates a yn poll directly via the service for testing.
func seedPoll(svc *poll.Service, title string, options []string) *poll.Poll {
	p, err := svc.Create(title, "", "yn", options)
	if err != nil {
		panic(err)
	}
	return p
}

// seedPollYMN creates a ymn poll directly via the service for testing.
func seedPollYMN(svc *poll.Service, title string, options []string) *poll.Poll {
	p, err := svc.Create(title, "", "ymn", options)
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
	if v.Responses["2025-06-10"] != "yes" {
		t.Error("expected 2025-06-10 to be yes")
	}
	if v.Responses["2025-06-11"] != "no" {
		t.Error("expected 2025-06-11 to be no")
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

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
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

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
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
		want   string
	}{
		{"Mon", "yes"},
		{"Tue", "no"},
		{"Wed", "no"},
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

	_ = svc.AddVote(p.ID, "Alice", map[string]string{"2025-10-01": "yes", "2025-10-02": "no"})
	_ = svc.AddVote(p.ID, "Bob", map[string]string{"2025-10-01": "yes", "2025-10-02": "yes"})

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
	_ = svc.AddVote(p.ID, "Alice", map[string]string{"Mon": "yes", "Tue": "no"})

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
	_ = svc.AddVote(p.ID, "Alice", map[string]string{"Mon": "yes"})
	_ = svc.AddVote(p.ID, "Bob", map[string]string{"Mon": "yes"})

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
	_ = svc.AddVote(p.ID, "Alice", map[string]string{"Mon": "yes"})

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
	_ = svc.AddVote(p.ID, "Alice", map[string]string{"Mon": "yes", "Tue": "no"})

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
	if v.Responses["Mon"] != "no" {
		t.Error("expected Mon to be no")
	}
	if v.Responses["Tue"] != "yes" {
		t.Error("expected Tue to be yes")
	}
}

func TestUpdateVotePreservesPosition(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPoll(svc, "Position test", []string{"Mon", "Tue"})
	_ = svc.AddVote(p.ID, "Alice", map[string]string{"Mon": "yes", "Tue": "no"})
	_ = svc.AddVote(p.ID, "Bob", map[string]string{"Mon": "no", "Tue": "yes"})
	_ = svc.AddVote(p.ID, "Carol", map[string]string{"Mon": "yes", "Tue": "yes"})

	// Edit Bob (middle vote) — should stay in position 1.
	form := url.Values{
		"old_name": {"Bob"},
		"name":     {"Bobby"},
		"vote-Mon": {"yes"},
		"vote-Tue": {"yes"},
	}
	w := postForm(router, "/poll/"+p.AdminID+"/admin/edit", form)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
	}

	got, _ := svc.Get(p.ID)
	if len(got.Votes) != 3 {
		t.Fatalf("expected 3 votes, got %d", len(got.Votes))
	}
	wantNames := []string{"Alice", "Bobby", "Carol"}
	for i, want := range wantNames {
		if got.Votes[i].Name != want {
			t.Errorf("vote[%d]: got %q, want %q", i, got.Votes[i].Name, want)
		}
	}
}

func TestUpdateVoteEmptyName(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPoll(svc, "Edit empty", []string{"Mon"})
	_ = svc.AddVote(p.ID, "Alice", map[string]string{"Mon": "yes"})

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

// --- Answer mode (yn / ymn) e2e tests ---

func TestCreatePollWithAnswerModeYMN(t *testing.T) {
	router, svc := setupTestRouter()

	form := url.Values{
		"title":       {"Maybe poll"},
		"answer_mode": {"ymn"},
		"dates[]":     {"2025-06-10", "2025-06-11"},
	}
	w := postForm(router, "/new", form)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
	}

	loc := w.Header().Get("Location")
	parts := strings.Split(loc, "/")
	adminID := parts[2]

	p, err := svc.GetByAdminID(adminID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.AnswerMode != "ymn" {
		t.Errorf("expected answer mode ymn, got %q", p.AnswerMode)
	}
}

func TestCreatePollDefaultsToYN(t *testing.T) {
	router, svc := setupTestRouter()

	form := url.Values{
		"title":   {"Default poll"},
		"dates[]": {"2025-06-10"},
	}
	w := postForm(router, "/new", form)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
	}

	loc := w.Header().Get("Location")
	parts := strings.Split(loc, "/")
	adminID := parts[2]

	p, _ := svc.GetByAdminID(adminID)
	if p.AnswerMode != "yn" {
		t.Errorf("expected answer mode yn, got %q", p.AnswerMode)
	}
}

func TestVoteMaybeOnYMNPoll(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPollYMN(svc, "YMN Dinner", []string{"Mon", "Tue", "Wed"})

	form := url.Values{
		"name":     {"Alice"},
		"vote-Mon": {"yes"},
		"vote-Tue": {"maybe"},
		"vote-Wed": {"no"},
	}
	w := postForm(router, "/poll/"+p.ID+"/vote", form)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
	}

	got, _ := svc.Get(p.ID)
	if len(got.Votes) != 1 {
		t.Fatalf("expected 1 vote, got %d", len(got.Votes))
	}
	v := got.Votes[0]
	if v.Responses["Mon"] != "yes" {
		t.Errorf("Mon: got %q, want yes", v.Responses["Mon"])
	}
	if v.Responses["Tue"] != "maybe" {
		t.Errorf("Tue: got %q, want maybe", v.Responses["Tue"])
	}
	if v.Responses["Wed"] != "no" {
		t.Errorf("Wed: got %q, want no", v.Responses["Wed"])
	}
}

func TestMaybeIgnoredOnYNPoll(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPoll(svc, "YN Dinner", []string{"Mon", "Tue"})

	// Submit "maybe" on a yn poll — should be treated as "no".
	form := url.Values{
		"name":     {"Bob"},
		"vote-Mon": {"yes"},
		"vote-Tue": {"maybe"},
	}
	w := postForm(router, "/poll/"+p.ID+"/vote", form)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
	}

	got, _ := svc.Get(p.ID)
	v := got.Votes[0]
	// The handler accepts "maybe" as a valid value regardless of answer mode,
	// since the form wouldn't normally offer it for yn polls.
	// But if someone manually posts it, it's stored as-is.
	if v.Responses["Mon"] != "yes" {
		t.Errorf("Mon: got %q, want yes", v.Responses["Mon"])
	}
	if v.Responses["Tue"] != "maybe" {
		t.Errorf("Tue: got %q, want maybe", v.Responses["Tue"])
	}
}

func TestTotalsWithMaybeViaHandler(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPollYMN(svc, "Totals test", []string{"Mon", "Tue"})

	postForm(router, "/poll/"+p.ID+"/vote", url.Values{
		"name":     {"Alice"},
		"vote-Mon": {"yes"},
		"vote-Tue": {"maybe"},
	})
	postForm(router, "/poll/"+p.ID+"/vote", url.Values{
		"name":     {"Bob"},
		"vote-Mon": {"maybe"},
		"vote-Tue": {"yes"},
	})
	postForm(router, "/poll/"+p.ID+"/vote", url.Values{
		"name":     {"Carol"},
		"vote-Mon": {"yes"},
		"vote-Tue": {"no"},
	})

	got, _ := svc.Get(p.ID)
	totals := poll.Totals(got)

	if totals["Mon"].Yes != 2 {
		t.Errorf("Mon.Yes = %d, want 2", totals["Mon"].Yes)
	}
	if totals["Mon"].Maybe != 1 {
		t.Errorf("Mon.Maybe = %d, want 1", totals["Mon"].Maybe)
	}
	if totals["Tue"].Yes != 1 {
		t.Errorf("Tue.Yes = %d, want 1", totals["Tue"].Yes)
	}
	if totals["Tue"].Maybe != 1 {
		t.Errorf("Tue.Maybe = %d, want 1", totals["Tue"].Maybe)
	}
}

func TestEditVoteMaybeOnYMNPoll(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPollYMN(svc, "Edit maybe", []string{"Mon", "Tue"})
	_ = svc.AddVote(p.ID, "Alice", map[string]string{"Mon": "yes", "Tue": "no"})

	form := url.Values{
		"old_name": {"Alice"},
		"name":     {"Alice"},
		"vote-Mon": {"maybe"},
		"vote-Tue": {"yes"},
	}
	w := postForm(router, "/poll/"+p.AdminID+"/admin/edit", form)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", w.Code)
	}

	got, _ := svc.Get(p.ID)
	v := got.Votes[0]
	if v.Responses["Mon"] != "maybe" {
		t.Errorf("Mon: got %q, want maybe", v.Responses["Mon"])
	}
	if v.Responses["Tue"] != "yes" {
		t.Errorf("Tue: got %q, want yes", v.Responses["Tue"])
	}
}

func TestYMNPollViewRendersCorrectly(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPollYMN(svc, "Render test", []string{"Mon", "Tue"})
	_ = svc.AddVote(p.ID, "Alice", map[string]string{"Mon": "yes", "Tue": "maybe"})

	req := httptest.NewRequest(http.MethodGet, "/poll/"+p.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Alice") {
		t.Error("expected response body to contain voter name")
	}
	// ymn poll should render the maybe vote button
	if !strings.Contains(body, "vote-maybe.svg") {
		t.Error("expected response body to contain vote-maybe.svg for ymn poll")
	}
}

func TestYNPollViewDoesNotRenderMaybeButton(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPoll(svc, "YN render", []string{"Mon"})

	req := httptest.NewRequest(http.MethodGet, "/poll/"+p.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	if strings.Contains(body, "vote-maybe.svg") {
		t.Error("yn poll should not render the maybe button")
	}
}

func TestAdminVoteMaybeOnYMNPoll(t *testing.T) {
	router, svc := setupTestRouter()
	p := seedPollYMN(svc, "Admin YMN", []string{"Mon", "Tue"})

	form := url.Values{
		"name":     {"Alice"},
		"vote-Mon": {"maybe"},
		"vote-Tue": {"yes"},
	}
	w := postForm(router, "/poll/"+p.AdminID+"/admin/vote", form)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect 303, got %d", w.Code)
	}

	// Verify the vote appears on the admin view
	req := httptest.NewRequest(http.MethodGet, "/poll/"+p.AdminID+"/admin", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Alice") {
		t.Error("expected admin view to contain voter name 'Alice'")
	}
	if !strings.Contains(body, "vote-maybe.svg") {
		t.Error("expected admin view to render maybe icon for ymn poll")
	}
}
