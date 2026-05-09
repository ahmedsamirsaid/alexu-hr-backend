package i18n

import (
	"context"
	"testing"
)

func TestService_T(t *testing.T) {
	// Setup test translations
	translations := map[string]map[string]string{
		"en": {
			"greeting":      "Hello",
			"audit.action.create": "Created",
		},
		"ar": {
			"greeting":      "مرحبا",
			"audit.action.create": "تم الإنشاء",
		},
	}

	service := NewService(translations, "en", []string{"en", "ar"})

	tests := []struct {
		name     string
		locale   string
		key      string
		expected string
	}{
		{
			name:     "translate to English from context",
			locale:   "en",
			key:      "greeting",
			expected: "Hello",
		},
		{
			name:     "translate to Arabic from context",
			locale:   "ar",
			key:      "greeting",
			expected: "مرحبا",
		},
		{
			name:     "translate audit action to English",
			locale:   "en",
			key:      "audit.action.create",
			expected: "Created",
		},
		{
			name:     "translate audit action to Arabic",
			locale:   "ar",
			key:      "audit.action.create",
			expected: "تم الإنشاء",
		},
		{
			name:     "missing key returns key itself",
			locale:   "en",
			key:      "missing.key",
			expected: "missing.key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithLocale(context.Background(), tt.locale)
			result := service.T(ctx, tt.key)
			if result != tt.expected {
				t.Errorf("T() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestService_TWithParams(t *testing.T) {
	translations := map[string]map[string]string{
		"en": {
			"welcome":       "Welcome, {{.Name}}!",
			"error.message": "Employee {{.EmployeeName}} not found",
		},
		"ar": {
			"welcome":       "مرحبا، {{.Name}}!",
			"error.message": "الموظف {{.EmployeeName}} غير موجود",
		},
	}

	service := NewService(translations, "en", []string{"en", "ar"})

	tests := []struct {
		name     string
		locale   string
		key      string
		params   map[string]interface{}
		expected string
	}{
		{
			name:   "substitute single parameter in English",
			locale: "en",
			key:    "welcome",
			params: map[string]interface{}{
				"Name": "Ahmed",
			},
			expected: "Welcome, Ahmed!",
		},
		{
			name:   "substitute single parameter in Arabic",
			locale: "ar",
			key:    "welcome",
			params: map[string]interface{}{
				"Name": "أحمد",
			},
			expected: "مرحبا، أحمد!",
		},
		{
			name:   "substitute parameter in error message",
			locale: "en",
			key:    "error.message",
			params: map[string]interface{}{
				"EmployeeName": "John Doe",
			},
			expected: "Employee John Doe not found",
		},
		{
			name:     "no parameters provided",
			locale:   "en",
			key:      "welcome",
			params:   map[string]interface{}{},
			expected: "Welcome, {{.Name}}!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := WithLocale(context.Background(), tt.locale)
			result := service.TWithParams(ctx, tt.key, tt.params)
			if result != tt.expected {
				t.Errorf("TWithParams() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestService_TLocale(t *testing.T) {
	translations := map[string]map[string]string{
		"en": {
			"notification.title": "Leave Approved",
		},
		"ar": {
			"notification.title": "تمت الموافقة على الإجازة",
		},
	}

	service := NewService(translations, "en", []string{"en", "ar"})

	tests := []struct {
		name     string
		locale   string
		key      string
		expected string
	}{
		{
			name:     "translate to English bypassing context",
			locale:   "en",
			key:      "notification.title",
			expected: "Leave Approved",
		},
		{
			name:     "translate to Arabic bypassing context",
			locale:   "ar",
			key:      "notification.title",
			expected: "تمت الموافقة على الإجازة",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TLocale should bypass context, so we don't need to pass context
			result := service.TLocale(tt.locale, tt.key)
			if result != tt.expected {
				t.Errorf("TLocale() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestService_TLocaleWithParams(t *testing.T) {
	translations := map[string]map[string]string{
		"en": {
			"notification.body": "Your leave from {{.StartDate}} to {{.EndDate}} has been approved",
		},
		"ar": {
			"notification.body": "تمت الموافقة على إجازتك من {{.StartDate}} إلى {{.EndDate}}",
		},
	}

	service := NewService(translations, "en", []string{"en", "ar"})

	tests := []struct {
		name     string
		locale   string
		key      string
		params   map[string]interface{}
		expected string
	}{
		{
			name:   "translate with params to English bypassing context",
			locale: "en",
			key:    "notification.body",
			params: map[string]interface{}{
				"StartDate": "2024-01-01",
				"EndDate":   "2024-01-05",
			},
			expected: "Your leave from 2024-01-01 to 2024-01-05 has been approved",
		},
		{
			name:   "translate with params to Arabic bypassing context",
			locale: "ar",
			key:    "notification.body",
			params: map[string]interface{}{
				"StartDate": "2024-01-01",
				"EndDate":   "2024-01-05",
			},
			expected: "تمت الموافقة على إجازتك من 2024-01-01 إلى 2024-01-05",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.TLocaleWithParams(tt.locale, tt.key, tt.params)
			if result != tt.expected {
				t.Errorf("TLocaleWithParams() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestService_GetSupportedLocales(t *testing.T) {
	translations := map[string]map[string]string{
		"en": {"key": "value"},
		"ar": {"key": "قيمة"},
	}

	service := NewService(translations, "en", []string{"en", "ar"})

	locales := service.GetSupportedLocales()
	if len(locales) != 2 {
		t.Errorf("GetSupportedLocales() returned %d locales, want 2", len(locales))
	}

	// Check that both locales are present
	localeMap := make(map[string]bool)
	for _, locale := range locales {
		localeMap[locale] = true
	}

	if !localeMap["en"] || !localeMap["ar"] {
		t.Errorf("GetSupportedLocales() = %v, want [en, ar]", locales)
	}
}

func TestService_GetDefaultLocale(t *testing.T) {
	translations := map[string]map[string]string{
		"en": {"key": "value"},
		"ar": {"key": "قيمة"},
	}

	service := NewService(translations, "en", []string{"en", "ar"})

	defaultLocale := service.GetDefaultLocale()
	if defaultLocale != "en" {
		t.Errorf("GetDefaultLocale() = %v, want en", defaultLocale)
	}
}

func TestService_ContextWithoutLocale(t *testing.T) {
	translations := map[string]map[string]string{
		"en": {
			"greeting": "Hello",
		},
		"ar": {
			"greeting": "مرحبا",
		},
	}

	service := NewService(translations, "en", []string{"en", "ar"})

	// Context without locale should use default locale
	ctx := context.Background()
	result := service.T(ctx, "greeting")
	if result != "Hello" {
		t.Errorf("T() with context without locale = %v, want Hello", result)
	}
}

func TestService_FallbackToDefaultLocale(t *testing.T) {
	translations := map[string]map[string]string{
		"en": {
			"greeting": "Hello",
		},
		"ar": {
			// Missing "greeting" key in Arabic
		},
	}

	service := NewService(translations, "en", []string{"en", "ar"})

	// Request Arabic translation, but key is missing, should fallback to English
	ctx := WithLocale(context.Background(), "ar")
	result := service.T(ctx, "greeting")
	if result != "Hello" {
		t.Errorf("T() with missing Arabic translation = %v, want Hello (fallback)", result)
	}
}
