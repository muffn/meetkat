package i18n

import "testing"

func TestNew(t *testing.T) {
	tr, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if tr == nil {
		t.Fatal("expected non-nil Translator")
	}
}

func TestBasicLookup(t *testing.T) {
	tr, _ := New()

	tests := []struct {
		lang string
		key  string
		want string
	}{
		{"en", "nav.back", "Back"},
		{"de", "nav.back", "Zurück"},
		{"en", "home.cta_start", "Get Started"},
		{"de", "home.cta_start", "Jetzt loslegen"},
	}

	for _, tt := range tests {
		loc := tr.ForLang(tt.lang)
		got := loc.T(tt.key)
		if got != tt.want {
			t.Errorf("ForLang(%q).T(%q) = %q, want %q", tt.lang, tt.key, got, tt.want)
		}
	}
}

func TestFormatStrings(t *testing.T) {
	tr, _ := New()

	tests := []struct {
		lang string
		key  string
		args []any
		want string
	}{
		{"en", "poll.page_title", []any{"Dinner"}, "Dinner – meetkat"},
		{"de", "poll.created_at", []any{"01.06.2025"}, "Erstellt am 01.06.2025"},
		{"en", "admin.remove_title", []any{"Alice"}, "Remove Alice"},
		{"de", "admin.remove_title", []any{"Alice"}, "Alice entfernen"},
	}

	for _, tt := range tests {
		loc := tr.ForLang(tt.lang)
		got := loc.T(tt.key, tt.args...)
		if got != tt.want {
			t.Errorf("ForLang(%q).T(%q, %v) = %q, want %q", tt.lang, tt.key, tt.args, got, tt.want)
		}
	}
}

func TestFallbackToEnglish(t *testing.T) {
	tr, _ := New()
	// German localizer looking up a key that only exists in English (if any).
	// Since we have full translations, test with an unknown language that falls back.
	loc := tr.ForLang("fr")
	got := loc.T("nav.back")
	if got != "Back" {
		t.Errorf("expected English fallback, got %q", got)
	}
}

func TestMissingKeyReturnsKey(t *testing.T) {
	tr, _ := New()
	loc := tr.ForLang("en")
	key := "nonexistent.key"
	got := loc.T(key)
	if got != key {
		t.Errorf("expected raw key %q, got %q", key, got)
	}
}

func TestLang(t *testing.T) {
	tr, _ := New()

	tests := []struct {
		input string
		want  string
	}{
		{"en", "en"},
		{"de", "de"},
		{"fr", "en"}, // falls back
	}

	for _, tt := range tests {
		loc := tr.ForLang(tt.input)
		if got := loc.Lang(); got != tt.want {
			t.Errorf("ForLang(%q).Lang() = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMatch(t *testing.T) {
	tr, _ := New()

	tests := []struct {
		accept string
		want   string
	}{
		{"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7", "de"},
		{"en-US,en;q=0.9", "en"},
		{"fr-FR,fr;q=0.9", "en"}, // no French, fallback
		{"", "en"},
		{"*", "en"}, // wildcard matches fallback (first supported)
	}

	for _, tt := range tests {
		got := tr.Match(tt.accept)
		if got != tt.want {
			t.Errorf("Match(%q) = %q, want %q", tt.accept, got, tt.want)
		}
	}
}
