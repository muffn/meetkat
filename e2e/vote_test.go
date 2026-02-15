//go:build e2e

package e2e

import (
	"testing"

	"github.com/chromedp/chromedp"
)

func TestVoteButtonToggle(t *testing.T) {
	ts := startTestServer(t)
	p := seedPoll(t, ts.Svc, "Toggle Test", []string{"2025-06-10", "2025-06-11"})
	ctx := newBrowserCtx(t)

	var hiddenVal string

	err := chromedp.Run(ctx,
		chromedp.Navigate(ts.URL+"/poll/"+p.ID),
		waitForSelector("#vote-name"),

		// Click the "yes" button for the first option in the inline vote row.
		// The inline vote row is the last <tr> in tbody. Its vote buttons have
		// hidden inputs with name="vote-2025-06-10".
		// We target the vote-btn inside the inline row (the row with #vote-name).
		chromedp.Click(`#vote-name`, chromedp.ByQuery),

		// Click yes for first option: find the td containing the hidden input for 2025-06-10,
		// then click its yes button.
		chromedp.Click(`input[name="vote-2025-06-10"] ~ div .vote-btn[data-value="yes"]`, chromedp.ByQuery),

		// Verify hidden input changed to "yes"
		chromedp.Value(`input[name="vote-2025-06-10"]`, &hiddenVal, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("vote button toggle (yes): %v", err)
	}
	if hiddenVal != "yes" {
		t.Errorf("expected hidden input 'yes', got %q", hiddenVal)
	}

	// Now click "no" for the same option and verify it switches.
	err = chromedp.Run(ctx,
		chromedp.Click(`input[name="vote-2025-06-10"] ~ div .vote-btn[data-value="no"]`, chromedp.ByQuery),
		chromedp.Value(`input[name="vote-2025-06-10"]`, &hiddenVal, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("vote button toggle (no): %v", err)
	}
	if hiddenVal != "no" {
		t.Errorf("expected hidden input 'no', got %q", hiddenVal)
	}

	// Verify the second option is still unset.
	var secondVal string
	err = chromedp.Run(ctx,
		chromedp.Value(`input[name="vote-2025-06-11"]`, &secondVal, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("read second option: %v", err)
	}
	if secondVal != "" {
		t.Errorf("expected second option to still be empty, got %q", secondVal)
	}
}

func TestAJAXVoteSubmission(t *testing.T) {
	ts := startTestServer(t)
	p := seedPoll(t, ts.Svc, "AJAX Vote", []string{"Mon", "Tue"})
	ctx := newBrowserCtx(t)

	var tableHTML string

	err := chromedp.Run(ctx,
		chromedp.Navigate(ts.URL+"/poll/"+p.ID),
		waitForSelector("#vote-name"),

		// Enter a name
		chromedp.SendKeys("#vote-name", "Alice", chromedp.ByQuery),

		// Click yes on both options
		chromedp.Click(`input[name="vote-Mon"] ~ div .vote-btn[data-value="yes"]`, chromedp.ByQuery),
		chromedp.Click(`input[name="vote-Tue"] ~ div .vote-btn[data-value="yes"]`, chromedp.ByQuery),

		// Submit the form
		chromedp.Click("#vote-submit", chromedp.ByQuery),

		// Wait for the AJAX response to swap the table content.
		// After swap, "Alice" should appear in the vote table as a voter row.
		chromedp.WaitVisible(`#vote-table-wrapper td`, chromedp.ByQuery),
		chromedp.Sleep(500*1e6), // 500ms for AJAX to complete

		// Read the table HTML to verify Alice appears.
		chromedp.InnerHTML("#vote-table-wrapper", &tableHTML, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("AJAX vote submission: %v", err)
	}

	if !contains(tableHTML, "Alice") {
		t.Error("expected 'Alice' to appear in the vote table after AJAX submission")
	}

	// Verify the vote was persisted server-side.
	got, _ := ts.Svc.Get(p.ID)
	if len(got.Votes) != 1 {
		t.Fatalf("expected 1 vote server-side, got %d", len(got.Votes))
	}
	if got.Votes[0].Name != "Alice" {
		t.Errorf("expected vote name 'Alice', got %q", got.Votes[0].Name)
	}
}

func TestTwoClickConfirmIncomplete(t *testing.T) {
	ts := startTestServer(t)
	p := seedPoll(t, ts.Svc, "Confirm Test", []string{"Mon", "Tue"})
	ctx := newBrowserCtx(t)

	var btnClasses string
	var confirmArmed string

	err := chromedp.Run(ctx,
		chromedp.Navigate(ts.URL+"/poll/"+p.ID),
		waitForSelector("#vote-name"),

		// Enter a name but leave all options unanswered.
		chromedp.SendKeys("#vote-name", "Bob", chromedp.ByQuery),

		// First click on submit — should arm the confirmation, NOT submit.
		chromedp.Click("#vote-submit", chromedp.ByQuery),
		chromedp.Sleep(200*1e6),

		// Check that the form is now in "armed" state.
		chromedp.AttributeValue(`form[data-confirm-incomplete]`, "data-confirm-armed", &confirmArmed, nil, chromedp.ByQuery),

		// Check that the button has the amber warning class.
		chromedp.AttributeValue("#vote-submit", "class", &btnClasses, nil, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("first click (arm): %v", err)
	}
	if confirmArmed != "true" {
		t.Errorf("expected data-confirm-armed='true', got %q", confirmArmed)
	}
	if !contains(btnClasses, "bg-amber-500") {
		t.Errorf("expected amber button class, got %q", btnClasses)
	}

	// Verify no vote was submitted yet.
	got, _ := ts.Svc.Get(p.ID)
	if len(got.Votes) != 0 {
		t.Fatalf("expected 0 votes after first click, got %d", len(got.Votes))
	}

	// Second click — should actually submit.
	err = chromedp.Run(ctx,
		chromedp.Click("#vote-submit", chromedp.ByQuery),
		chromedp.Sleep(500*1e6),
	)
	if err != nil {
		t.Fatalf("second click (submit): %v", err)
	}

	got, _ = ts.Svc.Get(p.ID)
	if len(got.Votes) != 1 {
		t.Fatalf("expected 1 vote after second click, got %d", len(got.Votes))
	}
	if got.Votes[0].Name != "Bob" {
		t.Errorf("expected vote name 'Bob', got %q", got.Votes[0].Name)
	}
}

func TestTwoClickConfirmResetOnVoteButton(t *testing.T) {
	ts := startTestServer(t)
	p := seedPoll(t, ts.Svc, "Reset Test", []string{"Mon", "Tue"})
	ctx := newBrowserCtx(t)

	var confirmArmed string

	err := chromedp.Run(ctx,
		chromedp.Navigate(ts.URL+"/poll/"+p.ID),
		waitForSelector("#vote-name"),

		// Enter name, leave options empty, click submit to arm.
		chromedp.SendKeys("#vote-name", "Carol", chromedp.ByQuery),
		chromedp.Click("#vote-submit", chromedp.ByQuery),
		chromedp.Sleep(200*1e6),

		// Now click a vote button — this should reset the armed state.
		chromedp.Click(`input[name="vote-Mon"] ~ div .vote-btn[data-value="yes"]`, chromedp.ByQuery),
		chromedp.Sleep(200*1e6),

		// Check that armed state was reset.
		chromedp.AttributeValue(`form[data-confirm-incomplete]`, "data-confirm-armed", &confirmArmed, nil, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("reset confirm on vote button click: %v", err)
	}
	// After reset, data-confirm-armed should be empty or not "true".
	if confirmArmed == "true" {
		t.Error("expected confirm-armed to be reset after clicking a vote button")
	}

	// Verify no vote was submitted.
	got, _ := ts.Svc.Get(p.ID)
	if len(got.Votes) != 0 {
		t.Errorf("expected 0 votes, got %d", len(got.Votes))
	}
}

// contains checks if substr is in s (simple helper to avoid importing strings).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
