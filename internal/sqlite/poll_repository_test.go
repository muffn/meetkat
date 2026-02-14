package sqlite

import (
	"testing"

	"meetkat/internal/poll"
)

func openTestDB(t *testing.T) *PollRepository {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewPollRepository(db)
}

func TestCreateAndGet(t *testing.T) {
	repo := openTestDB(t)

	p := &poll.Poll{
		ID:          "abc12345",
		AdminID:     "adm12345",
		Title:       "Dinner",
		Description: "Pick a day",
		Options:     []string{"Mon", "Tue", "Wed"},
	}

	if err := repo.Create(p); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByPublicID("abc12345")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil {
		t.Fatal("expected poll, got nil")
	}
	if got.Title != "Dinner" {
		t.Errorf("title: got %q, want %q", got.Title, "Dinner")
	}
	if got.Description != "Pick a day" {
		t.Errorf("description: got %q, want %q", got.Description, "Pick a day")
	}
	if got.AdminID != "adm12345" {
		t.Errorf("admin_id: got %q, want %q", got.AdminID, "adm12345")
	}
	if len(got.Options) != 3 {
		t.Fatalf("options: got %d, want 3", len(got.Options))
	}
	// Options should preserve order.
	for i, want := range []string{"Mon", "Tue", "Wed"} {
		if got.Options[i] != want {
			t.Errorf("option[%d]: got %q, want %q", i, got.Options[i], want)
		}
	}
	if len(got.Votes) != 0 {
		t.Errorf("votes: got %d, want 0", len(got.Votes))
	}
}

func TestGetNotFound(t *testing.T) {
	repo := openTestDB(t)

	got, err := repo.GetByPublicID("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil, got poll")
	}
}

func TestGetByAdminID(t *testing.T) {
	repo := openTestDB(t)

	p := &poll.Poll{
		ID:          "pub12345",
		AdminID:     "adm99999",
		Title:       "Admin test",
		Description: "",
		Options:     []string{"A"},
	}
	if err := repo.Create(p); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByAdminID("adm99999")
	if err != nil {
		t.Fatalf("get by admin id: %v", err)
	}
	if got == nil {
		t.Fatal("expected poll, got nil")
	}
	if got.ID != "pub12345" {
		t.Errorf("public id: got %q, want %q", got.ID, "pub12345")
	}
	if got.AdminID != "adm99999" {
		t.Errorf("admin id: got %q, want %q", got.AdminID, "adm99999")
	}
}

func TestGetByAdminIDNotFound(t *testing.T) {
	repo := openTestDB(t)

	got, err := repo.GetByAdminID("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil, got poll")
	}
}

func TestAddVote(t *testing.T) {
	repo := openTestDB(t)

	p := &poll.Poll{
		ID:      "vote1234",
		AdminID: "adm_vote",
		Title:   "Lunch",
		Options: []string{"Mon", "Tue"},
	}
	if err := repo.Create(p); err != nil {
		t.Fatalf("create: %v", err)
	}

	err := repo.AddVote("vote1234", poll.Vote{
		Name:      "Alice",
		Responses: map[string]bool{"Mon": true, "Tue": false},
	})
	if err != nil {
		t.Fatalf("add vote: %v", err)
	}

	got, err := repo.GetByPublicID("vote1234")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil {
		t.Fatal("expected poll, got nil")
	}
	if len(got.Votes) != 1 {
		t.Fatalf("votes: got %d, want 1", len(got.Votes))
	}
	v := got.Votes[0]
	if v.Name != "Alice" {
		t.Errorf("name: got %q, want Alice", v.Name)
	}
	if !v.Responses["Mon"] {
		t.Error("expected Mon to be true")
	}
	if v.Responses["Tue"] {
		t.Error("expected Tue to be false")
	}
}

func TestAddVoteNonexistentPoll(t *testing.T) {
	repo := openTestDB(t)

	err := repo.AddVote("nope", poll.Vote{Name: "Bob", Responses: map[string]bool{}})
	if err == nil {
		t.Fatal("expected error for nonexistent poll")
	}
}

func TestRemoveVote(t *testing.T) {
	repo := openTestDB(t)

	p := &poll.Poll{
		ID:      "rmvote12",
		AdminID: "adm_rmv1",
		Title:   "Remove test",
		Options: []string{"Mon", "Tue"},
	}
	if err := repo.Create(p); err != nil {
		t.Fatalf("create: %v", err)
	}

	_ = repo.AddVote("rmvote12", poll.Vote{Name: "Alice", Responses: map[string]bool{"Mon": true}})
	_ = repo.AddVote("rmvote12", poll.Vote{Name: "Bob", Responses: map[string]bool{"Tue": true}})

	if err := repo.RemoveVote("rmvote12", "Alice"); err != nil {
		t.Fatalf("remove vote: %v", err)
	}

	got, _ := repo.GetByPublicID("rmvote12")
	if len(got.Votes) != 1 {
		t.Fatalf("votes: got %d, want 1", len(got.Votes))
	}
	if got.Votes[0].Name != "Bob" {
		t.Errorf("remaining vote: got %q, want Bob", got.Votes[0].Name)
	}
}

func TestRemoveVoteNotFound(t *testing.T) {
	repo := openTestDB(t)

	p := &poll.Poll{
		ID:      "rmvnf123",
		AdminID: "adm_rmvn",
		Title:   "Remove NF",
		Options: []string{"A"},
	}
	if err := repo.Create(p); err != nil {
		t.Fatalf("create: %v", err)
	}

	err := repo.RemoveVote("rmvnf123", "Nobody")
	if err == nil {
		t.Fatal("expected error for nonexistent voter")
	}
}

func TestMultipleVotes(t *testing.T) {
	repo := openTestDB(t)

	p := &poll.Poll{
		ID:      "multi123",
		AdminID: "adm_mult",
		Title:   "Sprint",
		Options: []string{"A", "B"},
	}
	if err := repo.Create(p); err != nil {
		t.Fatalf("create: %v", err)
	}

	_ = repo.AddVote("multi123", poll.Vote{Name: "Alice", Responses: map[string]bool{"A": true, "B": true}})
	_ = repo.AddVote("multi123", poll.Vote{Name: "Bob", Responses: map[string]bool{"A": true, "B": false}})

	got, _ := repo.GetByPublicID("multi123")
	if len(got.Votes) != 2 {
		t.Fatalf("votes: got %d, want 2", len(got.Votes))
	}
	if got.Votes[0].Name != "Alice" {
		t.Errorf("first vote: got %q, want Alice", got.Votes[0].Name)
	}
	if got.Votes[1].Name != "Bob" {
		t.Errorf("second vote: got %q, want Bob", got.Votes[1].Name)
	}
}

func TestUpdateVotePreservesPosition(t *testing.T) {
	repo := openTestDB(t)

	p := &poll.Poll{
		ID:      "upd12345",
		AdminID: "adm_upd1",
		Title:   "Update test",
		Options: []string{"Mon", "Tue"},
	}
	if err := repo.Create(p); err != nil {
		t.Fatalf("create: %v", err)
	}

	_ = repo.AddVote("upd12345", poll.Vote{Name: "Alice", Responses: map[string]bool{"Mon": true, "Tue": false}})
	_ = repo.AddVote("upd12345", poll.Vote{Name: "Bob", Responses: map[string]bool{"Mon": false, "Tue": true}})
	_ = repo.AddVote("upd12345", poll.Vote{Name: "Carol", Responses: map[string]bool{"Mon": true, "Tue": true}})

	// Update Bob's name and responses.
	err := repo.UpdateVote("upd12345", "Bob", poll.Vote{
		Name:      "Bobby",
		Responses: map[string]bool{"Mon": true, "Tue": true},
	})
	if err != nil {
		t.Fatalf("update vote: %v", err)
	}

	got, _ := repo.GetByPublicID("upd12345")
	if len(got.Votes) != 3 {
		t.Fatalf("votes: got %d, want 3", len(got.Votes))
	}

	// Order must be preserved: Alice, Bobby (was Bob), Carol.
	wantNames := []string{"Alice", "Bobby", "Carol"}
	for i, want := range wantNames {
		if got.Votes[i].Name != want {
			t.Errorf("vote[%d]: got %q, want %q", i, got.Votes[i].Name, want)
		}
	}

	// Bobby's responses should be updated.
	if !got.Votes[1].Responses["Mon"] {
		t.Error("expected Bobby's Mon to be true")
	}
	if !got.Votes[1].Responses["Tue"] {
		t.Error("expected Bobby's Tue to be true")
	}
}

func TestUpdateVoteNotFound(t *testing.T) {
	repo := openTestDB(t)

	p := &poll.Poll{
		ID:      "updnf123",
		AdminID: "adm_unf1",
		Title:   "Update NF",
		Options: []string{"A"},
	}
	if err := repo.Create(p); err != nil {
		t.Fatalf("create: %v", err)
	}

	err := repo.UpdateVote("updnf123", "Nobody", poll.Vote{Name: "X", Responses: map[string]bool{"A": true}})
	if err == nil {
		t.Fatal("expected error for nonexistent voter")
	}
}

func TestServiceWithSQLiteRepository(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := NewPollRepository(db)
	svc := poll.NewService(repo)

	p, err := svc.Create("End-to-end", "Test full flow", []string{"X", "Y"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := svc.AddVote(p.ID, "Carol", map[string]bool{"X": true, "Y": false}); err != nil {
		t.Fatalf("add vote: %v", err)
	}

	got, err := svc.Get(p.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil {
		t.Fatal("expected poll")
	}
	if len(got.Votes) != 1 {
		t.Fatalf("votes: got %d, want 1", len(got.Votes))
	}
	if got.Votes[0].Name != "Carol" {
		t.Errorf("name: got %q, want Carol", got.Votes[0].Name)
	}

	totals := poll.Totals(got)
	if totals["X"] != 1 {
		t.Errorf("totals[X]: got %d, want 1", totals["X"])
	}
	if totals["Y"] != 0 {
		t.Errorf("totals[Y]: got %d, want 0", totals["Y"])
	}
}
