package poll

import (
	"strings"
	"testing"
)

func TestCreate(t *testing.T) {
	svc := NewService()
	p := svc.Create("Dinner", []string{"Mon", "Tue"})

	if p.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if len(p.ID) != 8 {
		t.Fatalf("expected ID length 8, got %d", len(p.ID))
	}
	if p.Title != "Dinner" {
		t.Errorf("expected title Dinner, got %q", p.Title)
	}
	if len(p.Options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(p.Options))
	}
}

func TestGet(t *testing.T) {
	svc := NewService()
	created := svc.Create("Lunch", []string{"Wed"})

	got, ok := svc.Get(created.ID)
	if !ok {
		t.Fatal("expected poll to be found")
	}
	if got.Title != "Lunch" {
		t.Errorf("expected title Lunch, got %q", got.Title)
	}
}

func TestGetNotFound(t *testing.T) {
	svc := NewService()
	_, ok := svc.Get("doesnotexist")
	if ok {
		t.Fatal("expected poll not to be found")
	}
}

func TestAddVote(t *testing.T) {
	svc := NewService()
	p := svc.Create("Offsite", []string{"Mon", "Tue"})

	err := svc.AddVote(p.ID, "Alice", map[string]bool{"Mon": true, "Tue": false})
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
	if !got.Votes[0].Responses["Mon"] {
		t.Error("expected Mon to be true")
	}
	if got.Votes[0].Responses["Tue"] {
		t.Error("expected Tue to be false")
	}
}

func TestAddVoteEmptyName(t *testing.T) {
	svc := NewService()
	p := svc.Create("Test", []string{"A"})

	err := svc.AddVote(p.ID, "", map[string]bool{"A": true})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if len(p.Votes) != 0 {
		t.Error("vote should not have been saved")
	}
}

func TestAddVoteNonexistentPoll(t *testing.T) {
	svc := NewService()

	err := svc.AddVote("nope", "Alice", map[string]bool{})
	if err == nil {
		t.Fatal("expected error for nonexistent poll")
	}
}

func TestTotals(t *testing.T) {
	p := &Poll{
		Options: []string{"Mon", "Tue", "Wed"},
		Votes: []Vote{
			{Name: "Alice", Responses: map[string]bool{"Mon": true, "Tue": true, "Wed": false}},
			{Name: "Bob", Responses: map[string]bool{"Mon": true, "Tue": false, "Wed": true}},
		},
	}

	totals := Totals(p)
	tests := []struct {
		option string
		want   int
	}{
		{"Mon", 2},
		{"Tue", 1},
		{"Wed", 1},
	}
	for _, tt := range tests {
		if got := totals[tt.option]; got != tt.want {
			t.Errorf("Totals[%q] = %d, want %d", tt.option, got, tt.want)
		}
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
