//go:build e2e

package e2e

import (
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestAdminRemoveVote(t *testing.T) {
	ts := startTestServer(t)
	p := seedPoll(t, ts.Svc, "Remove Test", []string{"Mon", "Tue"})
	_ = ts.Svc.AddVote(p.ID, "Alice", map[string]bool{"Mon": true, "Tue": false})
	_ = ts.Svc.AddVote(p.ID, "Bob", map[string]bool{"Mon": true, "Tue": true})
	ctx := newBrowserCtx(t)

	var tableHTML string

	err := chromedp.Run(ctx,
		chromedp.Navigate(ts.URL+"/poll/"+p.AdminID+"/admin"),
		waitForSelector("#vote-table-wrapper"),

		// Click the remove button for Alice (data-action="remove" data-voter="Alice").
		chromedp.Click(`button[data-action="remove"][data-voter="Alice"]`, chromedp.ByQuery),

		// Wait for AJAX swap.
		chromedp.Sleep(500*1e6),

		// Read the table to verify Alice is gone and Bob remains.
		chromedp.InnerHTML("#vote-table-wrapper", &tableHTML, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("admin remove vote: %v", err)
	}

	if contains(tableHTML, "Alice") {
		t.Error("expected Alice to be removed from the table")
	}
	if !contains(tableHTML, "Bob") {
		t.Error("expected Bob to still be in the table")
	}

	// Verify server-side.
	got, _ := ts.Svc.Get(p.ID)
	if len(got.Votes) != 1 {
		t.Fatalf("expected 1 vote, got %d", len(got.Votes))
	}
	if got.Votes[0].Name != "Bob" {
		t.Errorf("expected remaining vote to be Bob, got %q", got.Votes[0].Name)
	}
}

func TestAdminEditVote(t *testing.T) {
	ts := startTestServer(t)
	p := seedPoll(t, ts.Svc, "Edit Test", []string{"Mon", "Tue"})
	_ = ts.Svc.AddVote(p.ID, "Alice", map[string]bool{"Mon": true, "Tue": false})
	ctx := newBrowserCtx(t)

	var tableHTML string

	// Step 1: Navigate and click edit button.
	err := chromedp.Run(ctx,
		chromedp.Navigate(ts.URL+"/poll/"+p.AdminID+"/admin"),
		waitForSelector("#vote-table-wrapper"),
	)
	if err != nil {
		t.Fatalf("navigate: %v", err)
	}

	// Step 2: Click the pencil button via JS to avoid selector issues.
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`startEdit(0)`, nil),
		chromedp.Sleep(200*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("startEdit: %v", err)
	}

	// Step 3: Clear name and type new name.
	err = chromedp.Run(ctx,
		chromedp.Clear(`#edit-0 input[name="name"]`, chromedp.ByQuery),
		chromedp.SendKeys(`#edit-0 input[name="name"]`, "Alicia", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("edit name: %v", err)
	}

	// Step 4: Toggle Tue to "yes" via JS (set hidden input directly + click button).
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`document.querySelector('#edit-0 input[name="vote-Tue"]').value = "yes"`, nil),
	)
	if err != nil {
		t.Fatalf("set Tue to yes: %v", err)
	}

	// Step 5: Click save.
	err = chromedp.Run(ctx,
		chromedp.Click(`button[data-action="edit-save"][data-idx="0"]`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.InnerHTML("#vote-table-wrapper", &tableHTML, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("save edit: %v", err)
	}

	if !contains(tableHTML, "Alicia") {
		t.Error("expected 'Alicia' to appear in the table after edit")
	}

	// Verify server-side.
	got, _ := ts.Svc.Get(p.ID)
	if len(got.Votes) != 1 {
		t.Fatalf("expected 1 vote, got %d", len(got.Votes))
	}
	if got.Votes[0].Name != "Alicia" {
		t.Errorf("expected name 'Alicia', got %q", got.Votes[0].Name)
	}
	if !got.Votes[0].Responses["Tue"] {
		t.Error("expected Tue to be true after edit")
	}
}
