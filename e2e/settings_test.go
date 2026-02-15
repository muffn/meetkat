//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// waitForClass polls until the element's class attribute contains (or no longer contains) the target string.
func waitForClass(sel, class string, shouldContain bool) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		deadline := time.After(5 * time.Second)
		for {
			var classes string
			if err := chromedp.AttributeValue(sel, "class", &classes, nil, chromedp.ByQuery).Do(ctx); err != nil {
				return err
			}
			if contains(classes, class) == shouldContain {
				return nil
			}
			select {
			case <-deadline:
				if shouldContain {
					return fmt.Errorf("timeout: %q never got class %q (last: %q)", sel, class, classes)
				}
				return fmt.Errorf("timeout: %q still has class %q (last: %q)", sel, class, classes)
			case <-time.After(50 * time.Millisecond):
			}
		}
	}
}

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

	// Click settings button to open, then wait for transition to complete.
	err = chromedp.Run(ctx,
		chromedp.Click("#settings-btn", chromedp.ByQuery),
		waitForClass("#settings-popover", "opacity-100", true),
	)
	if err != nil {
		t.Fatalf("open popover: %v", err)
	}

	// Click outside to close, then wait for transition to complete.
	err = chromedp.Run(ctx,
		chromedp.Click("body", chromedp.ByQuery),
		waitForClass("#settings-popover", "pointer-events-none", true),
	)
	if err != nil {
		t.Fatalf("close popover: %v", err)
	}
}

func TestThemeSwitch(t *testing.T) {
	ts := startTestServer(t)
	ctx := newBrowserCtx(t)

	var htmlClasses string

	err := chromedp.Run(ctx,
		chromedp.Navigate(ts.URL+"/"),
		waitForSelector("#settings-btn"),

		// Open settings and wait for popover.
		chromedp.Click("#settings-btn", chromedp.ByQuery),
		waitForClass("#settings-popover", "opacity-100", true),

		// Click the dark theme button.
		chromedp.Click(`button.theme-btn[data-theme="dark"]`, chromedp.ByQuery),

		// Wait for dark class on <html>.
		waitForClass("html", "dark", true),

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
		waitForClass("#settings-popover", "opacity-100", true),

		chromedp.Click(`button.theme-btn[data-theme="light"]`, chromedp.ByQuery),

		// Wait for dark class to be removed.
		waitForClass("html", "dark", false),

		chromedp.AttributeValue("html", "class", &htmlClasses, nil, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("theme switch to light: %v", err)
	}
	if contains(htmlClasses, "dark") {
		t.Errorf("expected <html> to NOT have 'dark' class after switching to light, got %q", htmlClasses)
	}
}
