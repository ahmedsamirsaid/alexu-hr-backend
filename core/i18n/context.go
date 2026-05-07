package i18n

import (
	"context"
	"sort"
	"strconv"
	"strings"
)

// contextKey is a private type for context keys to avoid collisions.
type contextKey string

const (
	// localeContextKey is the key used to store locale in context.
	localeContextKey contextKey = "locale"
	
	// defaultLocale is the fallback locale when none is specified.
	defaultLocale = "en"
)

// WithLocale injects a locale into the context.
func WithLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, localeContextKey, locale)
}

// LocaleFromContext extracts the locale from context.
// Returns default locale if none is set.
func LocaleFromContext(ctx context.Context) string {
	if locale, ok := ctx.Value(localeContextKey).(string); ok && locale != "" {
		return locale
	}
	return defaultLocale
}

// ParseAcceptLanguage parses Accept-Language header per RFC 7231.
// Returns the highest priority supported locale.
// 
// Example headers:
//   - "ar" -> "ar"
//   - "ar,en;q=0.9" -> "ar"
//   - "en-US,en;q=0.9,ar;q=0.8" -> "en"
//   - "fr,de;q=0.9" -> "en" (fallback if no supported locale found)
func ParseAcceptLanguage(header string, supported []string) string {
	if header == "" {
		return defaultLocale
	}

	// Create a map for quick lookup of supported locales
	supportedMap := make(map[string]bool)
	for _, locale := range supported {
		supportedMap[strings.ToLower(locale)] = true
	}

	// Parse language preferences with quality values
	type langPref struct {
		lang    string
		quality float64
	}

	var preferences []langPref

	// Split by comma to get individual language preferences
	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by semicolon to separate language from quality value
		langParts := strings.Split(part, ";")
		lang := strings.TrimSpace(langParts[0])
		
		// Extract base language code (e.g., "en" from "en-US")
		if idx := strings.Index(lang, "-"); idx != -1 {
			lang = lang[:idx]
		}
		lang = strings.ToLower(lang)

		// Parse quality value (default is 1.0)
		quality := 1.0
		if len(langParts) > 1 {
			qPart := strings.TrimSpace(langParts[1])
			if strings.HasPrefix(qPart, "q=") {
				if q, err := strconv.ParseFloat(qPart[2:], 64); err == nil {
					quality = q
				}
			}
		}

		preferences = append(preferences, langPref{lang: lang, quality: quality})
	}

	// Sort by quality value (descending)
	sort.Slice(preferences, func(i, j int) bool {
		return preferences[i].quality > preferences[j].quality
	})

	// Find the first supported language
	for _, pref := range preferences {
		if supportedMap[pref.lang] {
			return pref.lang
		}
	}

	// Fallback to default locale if no supported language found
	return defaultLocale
}
