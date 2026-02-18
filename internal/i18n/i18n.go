package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/text/language"
)

//go:embed translations/*.json
var translationFiles embed.FS

// Translator holds all loaded languages and a matcher for Accept-Language negotiation.
type Translator struct {
	messages  map[string]map[string]string // lang -> key -> message
	matcher   language.Matcher
	supported []string // lang codes in matcher order
	fallback  string
}

// Localizer provides translations for a single language.
type Localizer struct {
	lang     string
	messages map[string]string
	fallback map[string]string
}

// New loads all embedded translation JSON files and returns a Translator.
func New() (*Translator, error) {
	entries, err := translationFiles.ReadDir("translations")
	if err != nil {
		return nil, fmt.Errorf("read translations dir: %w", err)
	}

	const fallback = "en"
	messages := make(map[string]map[string]string)
	var langs []string

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		lang := strings.TrimSuffix(e.Name(), ".json")

		data, err := translationFiles.ReadFile("translations/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}

		var m map[string]string
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}

		messages[lang] = m
		langs = append(langs, lang)
	}

	// Ensure fallback language is first so the matcher uses it as default.
	ordered := make([]string, 0, len(langs))
	ordered = append(ordered, fallback)
	for _, l := range langs {
		if l != fallback {
			ordered = append(ordered, l)
		}
	}

	tags := make([]language.Tag, len(ordered))
	for i, l := range ordered {
		tag, err := language.Parse(l)
		if err != nil {
			return nil, fmt.Errorf("parse language tag %q: %w", l, err)
		}
		tags[i] = tag
	}

	return &Translator{
		messages:  messages,
		matcher:   language.NewMatcher(tags),
		supported: ordered,
		fallback:  fallback,
	}, nil
}

// IsSupported reports whether lang is a known supported language code.
func (t *Translator) IsSupported(lang string) bool {
	for _, s := range t.supported {
		if s == lang {
			return true
		}
	}
	return false
}

// ForLang returns a Localizer for the best-matching language.
func (t *Translator) ForLang(lang string) *Localizer {
	tag, err := language.Parse(lang)
	if err != nil {
		return t.localizer(t.fallback)
	}
	_, idx, _ := t.matcher.Match(tag)
	if idx >= 0 && idx < len(t.supported) {
		return t.localizer(t.supported[idx])
	}
	return t.localizer(t.fallback)
}

// Match parses an Accept-Language header and returns the best matching language code.
func (t *Translator) Match(acceptLanguage string) string {
	tags, _, err := language.ParseAcceptLanguage(acceptLanguage)
	if err != nil || len(tags) == 0 {
		return t.fallback
	}
	_, idx, _ := t.matcher.Match(tags...)
	if idx >= 0 && idx < len(t.supported) {
		return t.supported[idx]
	}
	return t.fallback
}

func (t *Translator) localizer(lang string) *Localizer {
	msgs := t.messages[lang]
	if msgs == nil {
		msgs = t.messages[t.fallback]
	}
	return &Localizer{
		lang:     lang,
		messages: msgs,
		fallback: t.messages[t.fallback],
	}
}

// T looks up a translation key. If args are provided, fmt.Sprintf is used.
// Falls back to English if the key is missing, then to the raw key.
func (l *Localizer) T(key string, args ...any) string {
	msg, ok := l.messages[key]
	if !ok {
		msg, ok = l.fallback[key]
		if !ok {
			return key
		}
	}
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

// Lang returns the language code for this localizer (e.g. "en", "de").
func (l *Localizer) Lang() string {
	return l.lang
}
