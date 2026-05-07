package i18n

import (
	"bytes"
	"html"
	"log/slog"
	"sync"
	"text/template"
)

// Localizer performs translation lookups from in-memory cache.
// It provides O(1) hash map lookups with concurrent access support.
type Localizer struct {
	// translations maps locale -> key -> value
	// Example: "en" -> "audit.action.create" -> "Created"
	translations map[string]map[string]string
	
	// defaultLocale is the fallback locale when translation is missing
	defaultLocale string
	
	// mu protects concurrent access to translations map
	mu sync.RWMutex
}

// NewLocalizer creates a new Localizer with the given translations.
func NewLocalizer(translations map[string]map[string]string, defaultLocale string) *Localizer {
	return &Localizer{
		translations:  translations,
		defaultLocale: defaultLocale,
	}
}

// Translate looks up a translation key for the given locale.
// Returns the original key if translation is missing (fallback behavior).
// Performs O(1) hash map lookup.
func (l *Localizer) Translate(locale, key string) string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Try requested locale first
	if localeMap, ok := l.translations[locale]; ok {
		if value, ok := localeMap[key]; ok {
			return value
		}
	}

	// Try default locale as fallback
	if locale != l.defaultLocale {
		if localeMap, ok := l.translations[l.defaultLocale]; ok {
			if value, ok := localeMap[key]; ok {
				return value
			}
		}
	}

	// Log missing translation
	slog.Warn("i18n.localizer.translation_missing",
		"locale", locale,
		"key", key,
	)

	// Return original key as final fallback
	return key
}

// TranslateWithParams performs translation with parameter substitution.
// Parameters are substituted using Go template syntax: {{.ParamName}}
// Parameter values are HTML-escaped to prevent injection attacks.
// Returns the original key if translation is missing.
func (l *Localizer) TranslateWithParams(locale, key string, params map[string]interface{}) string {
	// Get the translation template
	templateStr := l.Translate(locale, key)

	// If no parameters provided, return as-is
	if len(params) == 0 {
		return templateStr
	}

	// Escape HTML special characters in parameter values
	escapedParams := make(map[string]interface{})
	for k, v := range params {
		if str, ok := v.(string); ok {
			escapedParams[k] = html.EscapeString(str)
		} else {
			escapedParams[k] = v
		}
	}

	// Parse and execute template
	tmpl, err := template.New(key).Parse(templateStr)
	if err != nil {
		slog.Error("i18n.localizer.template_parse_error",
			"locale", locale,
			"key", key,
			"error", err,
		)
		return templateStr
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, escapedParams); err != nil {
		slog.Error("i18n.localizer.template_execute_error",
			"locale", locale,
			"key", key,
			"error", err,
		)
		return templateStr
	}

	return buf.String()
}

// GetSupportedLocales returns a list of all available locales.
func (l *Localizer) GetSupportedLocales() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	locales := make([]string, 0, len(l.translations))
	for locale := range l.translations {
		locales = append(locales, locale)
	}
	return locales
}

// GetDefaultLocale returns the fallback locale.
func (l *Localizer) GetDefaultLocale() string {
	return l.defaultLocale
}
