package i18n

import (
	"context"
	"testing"
)

func TestWithLocaleAndLocaleFromContext(t *testing.T) {
	tests := []struct {
		name     string
		locale   string
		expected string
	}{
		{
			name:     "set and retrieve English locale",
			locale:   "en",
			expected: "en",
		},
		{
			name:     "set and retrieve Arabic locale",
			locale:   "ar",
			expected: "ar",
		},
		{
			name:     "empty locale returns default",
			locale:   "",
			expected: "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			
			if tt.locale != "" {
				ctx = WithLocale(ctx, tt.locale)
			}
			
			got := LocaleFromContext(ctx)
			if got != tt.expected {
				t.Errorf("LocaleFromContext() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLocaleFromContext_NoLocaleSet(t *testing.T) {
	ctx := context.Background()
	got := LocaleFromContext(ctx)
	
	if got != "en" {
		t.Errorf("LocaleFromContext() with no locale set = %v, want %v", got, "en")
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	supported := []string{"en", "ar"}

	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "single language - Arabic",
			header:   "ar",
			expected: "ar",
		},
		{
			name:     "single language - English",
			header:   "en",
			expected: "en",
		},
		{
			name:     "multiple languages with quality - Arabic preferred",
			header:   "ar,en;q=0.9",
			expected: "ar",
		},
		{
			name:     "multiple languages with quality - English preferred",
			header:   "en,ar;q=0.9",
			expected: "en",
		},
		{
			name:     "language with region code",
			header:   "en-US,en;q=0.9,ar;q=0.8",
			expected: "en",
		},
		{
			name:     "Arabic with region code",
			header:   "ar-SA,ar;q=0.9",
			expected: "ar",
		},
		{
			name:     "unsupported language falls back to default",
			header:   "fr,de;q=0.9",
			expected: "en",
		},
		{
			name:     "empty header returns default",
			header:   "",
			expected: "en",
		},
		{
			name:     "complex quality values",
			header:   "fr;q=0.9,ar;q=0.8,en;q=0.7",
			expected: "ar",
		},
		{
			name:     "whitespace handling",
			header:   " ar , en ; q=0.9 ",
			expected: "ar",
		},
		{
			name:     "case insensitive",
			header:   "AR,EN;q=0.9",
			expected: "ar",
		},
		{
			name:     "quality value 0 should be ignored",
			header:   "fr;q=0,ar;q=0.5,en;q=0.3",
			expected: "ar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseAcceptLanguage(tt.header, supported)
			if got != tt.expected {
				t.Errorf("ParseAcceptLanguage(%q) = %v, want %v", tt.header, got, tt.expected)
			}
		})
	}
}

func TestParseAcceptLanguage_DifferentSupportedLocales(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		supported []string
		expected  string
	}{
		{
			name:      "only English supported",
			header:    "ar,en;q=0.9",
			supported: []string{"en"},
			expected:  "en",
		},
		{
			name:      "only Arabic supported",
			header:    "ar,en;q=0.9",
			supported: []string{"ar"},
			expected:  "ar",
		},
		{
			name:      "three languages supported",
			header:    "fr,ar;q=0.9,en;q=0.8",
			supported: []string{"en", "ar", "fr"},
			expected:  "fr",
		},
		{
			name:      "no supported languages match",
			header:    "de,fr;q=0.9",
			supported: []string{"en", "ar"},
			expected:  "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseAcceptLanguage(tt.header, tt.supported)
			if got != tt.expected {
				t.Errorf("ParseAcceptLanguage(%q, %v) = %v, want %v", tt.header, tt.supported, got, tt.expected)
			}
		})
	}
}
