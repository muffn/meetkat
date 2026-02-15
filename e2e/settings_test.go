//go:build e2e

package e2e

import (
	"testing"

	"github.com/chromedp/chromedp"
)

func TestSettingsPopoverToggle(t *testing.T) {
	ts := startTestServer(t)
	ctx := newBrowserCtx(t)

	var classes string

	err := chromedp.Run(ctx,
		chromedp.Navigate(ts.URL+"/"),
		waitForSelector("#settings-btn"),

		// Popover should start hidden (has pointer-events-none class).
		chromedp.AttributeValue("#settings-popover", "class", &classes, nil, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("initial state: %v", err)
	}
	if !contains(classes, "pointer-events-none") {
		t.Error("expected popover to start hidden (pointer-events-none)")
	}

	// Click settings button to open.
	err = chromedp.Run(ctx,
		chromedp.Click("#settings-btn", chromedp.ByQuery),
		chromedp.Sleep(300*1e6),
		chromedp.AttributeValue("#settings-popover", "class", &classes, nil, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("open popover: %v", err)
	}
	if contains(classes, "pointer-events-none") {
		t.Error("expected popover to be visible after clicking settings button")
	}
	if !contains(classes, "opacity-100") {
		t.Error("expected popover to have opacity-100 when open")
	}

	// Click outside to close.
	err = chromedp.Run(ctx,
		chromedp.Click("body", chromedp.ByQuery),
		chromedp.Sleep(300*1e6),
		chromedp.AttributeValue("#settings-popover", "class", &classes, nil, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("close popover: %v", err)
	}
	if !contains(classes, "pointer-events-none") {
		t.Error("expected popover to be hidden after clicking outside")
	}
}

func TestThemeSwitch(t *testing.T) {
	ts := startTestServer(t)
	ctx := newBrowserCtx(t)

	var htmlClasses string

	err := chromedp.Run(ctx,
		chromedp.Navigate(ts.URL+"/"),
		waitForSelector("#settings-btn"),

		// Open settings.
		chromedp.Click("#settings-btn", chromedp.ByQuery),
		chromedp.Sleep(300*1e6),

		// Click the dark theme button.
		chromedp.Click(`button.theme-btn[data-theme="dark"]`, chromedp.ByQuery),
		chromedp.Sleep(300*1e6),

		// Check that <html> has class "dark".
		chromedp.AttributeValue("html", "class", &htmlClasses, nil, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("theme switch to dark: %v", err)
	}
	if !contains(htmlClasses, "dark") {
		t.Errorf("expected <html> to have 'dark' class, got %q", htmlClasses)
	}

	// Switch back to light.
	err = chromedp.Run(ctx,
		chromedp.Click("#settings-btn", chromedp.ByQuery),
		chromedp.Sleep(300*1e6),
		chromedp.Click(`button.theme-btn[data-theme="light"]`, chromedp.ByQuery),
		chromedp.Sleep(300*1e6),
		chromedp.AttributeValue("html", "class", &htmlClasses, nil, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("theme switch to light: %v", err)
	}
	if contains(htmlClasses, "dark") {
		t.Errorf("expected <html> to NOT have 'dark' class after switching to light, got %q", htmlClasses)
	}
}
