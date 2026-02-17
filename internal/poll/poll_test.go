package poll

import (
	"strings"
	"testing"
)

func TestCreate(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	p, err := svc.Create("Dinner", "Pick your evening", "yn", []string{"Mon", "Tue"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if len(p.ID) != 8 {
		t.Fatalf("expected ID length 8, got %d", len(p.ID))
	}
	if p.AdminID == "" {
		t.Fatal("expected non-empty AdminID")
	}
	if len(p.AdminID) != 8 {
		t.Fatalf("expected AdminID length 8, got %d", len(p.AdminID))
	}
	if p.ID == p.AdminID {
		t.Error("expected ID and AdminID to be different")
	}
	if p.Title != "Dinner" {
		t.Errorf("expected title Dinner, got %q", p.Title)
	}
	if p.Description != "Pick your evening" {
		t.Errorf("expected description %q, got %q", "Pick your evening", p.Description)
	}
	if p.AnswerMode != "yn" {
		t.Errorf("expected answer mode yn, got %q", p.AnswerMode)
	}
	if len(p.Options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(p.Options))
	}
}

func TestCreateDefaultAnswerMode(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	p, err := svc.Create("Test", "", "invalid", []string{"A"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.AnswerMode != "yn" {
		t.Errorf("expected default answer mode yn, got %q", p.AnswerMode)
	}
}

func TestGet(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	created, err := svc.Create("Lunch", "", "yn", []string{"Wed"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := svc.Get(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected poll to be found")
	}
	if got.Title != "Lunch" {
		t.Errorf("expected title Lunch, got %q", got.Title)
	}
}

func TestGetNotFound(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	got, err := svc.Get("doesnotexist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatal("expected poll not to be found")
	}
}

func TestAddVote(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	p, err := svc.Create("Offsite", "", "yn", []string{"Mon", "Tue"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = svc.AddVote(p.ID, "Alice", map[string]string{"Mon": "yes", "Tue": "no"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := svc.Get(p.ID)
	if len(got.Votes) != 1 {
		t.Fatalf("expected 1 vote, got %d", len(got.Votes))
	}
	if got.Votes[0].Name != "Alice" {
		t.Errorf("expected name Alice, got %q", got.Votes[0].Name)
	}
	if got.Votes[0].Responses["Mon"] != "yes" {
		t.Error("expected Mon to be yes")
	}
	if got.Votes[0].Responses["Tue"] != "no" {
		t.Error("expected Tue to be no")
	}
}

func TestAddVoteEmptyName(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	p, _ := svc.Create("Test", "", "yn", []string{"A"})

	err := svc.AddVote(p.ID, "", map[string]string{"A": "yes"})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if len(p.Votes) != 0 {
		t.Error("vote should not have been saved")
	}
}

func TestAddVoteNonexistentPoll(t *testing.T) {
	svc := NewService(NewMemoryRepository())

	err := svc.AddVote("nope", "Alice", map[string]string{})
	if err == nil {
		t.Fatal("expected error for nonexistent poll")
	}
}

func TestTotals(t *testing.T) {
	p := &Poll{
		Options: []string{"Mon", "Tue", "Wed"},
		Votes: []Vote{
			{Name: "Alice", Responses: map[string]string{"Mon": "yes", "Tue": "yes", "Wed": "no"}},
			{Name: "Bob", Responses: map[string]string{"Mon": "yes", "Tue": "no", "Wed": "yes"}},
		},
	}

	totals := Totals(p)
	tests := []struct {
		option  string
		wantYes int
	}{
		{"Mon", 2},
		{"Tue", 1},
		{"Wed", 1},
	}
	for _, tt := range tests {
		if got := totals[tt.option].Yes; got != tt.wantYes {
			t.Errorf("Totals[%q].Yes = %d, want %d", tt.option, got, tt.wantYes)
		}
	}
}

func TestTotalsWithMaybe(t *testing.T) {
	p := &Poll{
		AnswerMode: "ymn",
		Options:    []string{"Mon", "Tue"},
		Votes: []Vote{
			{Name: "Alice", Responses: map[string]string{"Mon": "yes", "Tue": "maybe"}},
			{Name: "Bob", Responses: map[string]string{"Mon": "maybe", "Tue": "yes"}},
			{Name: "Carol", Responses: map[string]string{"Mon": "yes", "Tue": "no"}},
		},
	}

	totals := Totals(p)
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

func TestGetByAdminID(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	created, err := svc.Create("Meeting", "", "yn", []string{"Mon"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := svc.GetByAdminID(created.AdminID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected poll to be found")
	}
	if got.Title != "Meeting" {
		t.Errorf("expected title Meeting, got %q", got.Title)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, got.ID)
	}
}

func TestGetByAdminIDNotFound(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	got, err := svc.GetByAdminID("doesnotexist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatal("expected poll not to be found")
	}
}

func TestRemoveVote(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	p, err := svc.Create("Offsite", "", "yn", []string{"Mon", "Tue"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = svc.AddVote(p.ID, "Alice", map[string]string{"Mon": "yes"})
	_ = svc.AddVote(p.ID, "Bob", map[string]string{"Tue": "yes"})

	if err := svc.RemoveVote(p.ID, "Alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := svc.Get(p.ID)
	if len(got.Votes) != 1 {
		t.Fatalf("expected 1 vote, got %d", len(got.Votes))
	}
	if got.Votes[0].Name != "Bob" {
		t.Errorf("expected remaining vote to be Bob, got %q", got.Votes[0].Name)
	}
}

func TestRemoveVoteNotFound(t *testing.T) {
	svc := NewService(NewMemoryRepository())
	p, _ := svc.Create("Test", "", "yn", []string{"A"})
	_ = svc.AddVote(p.ID, "Alice", map[string]string{"A": "yes"})

	err := svc.RemoveVote(p.ID, "Nobody")
	if err == nil {
		t.Fatal("expected error for nonexistent voter")
	}
}

func TestGenerateID(t *testing.T) {
	id := generateID()
	if len(id) != 8 {
		t.Fatalf("expected ID length 8, got %d", len(id))
	}
	for _, c := range id {
		if !strings.ContainsRune(idChars, c) {
			t.Errorf("unexpected character %q in ID", c)
		}
	}
}
