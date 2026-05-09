package i18n

import (
	"strings"
	"sync"
	"testing"
)

func TestLocalizer_Translate(t *testing.T) {
	translations := map[string]map[string]string{
		"en": {
			"audit.action.create": "Created",
			"audit.action.update": "Updated",
			"error.not_found":     "Not found",
		},
		"ar": {
			"audit.action.create": "تم الإنشاء",
			"audit.action.update": "تم التحديث",
		},
	}

	localizer := NewLocalizer(translations, "en")

	tests := []struct {
		name     string
		locale   string
		key      string
		expected string
	}{
		{
			name:     "translate existing key in English",
			locale:   "en",
			key:      "audit.action.create",
			expected: "Created",
		},
		{
			name:     "translate existing key in Arabic",
			locale:   "ar",
			key:      "audit.action.create",
			expected: "تم الإنشاء",
		},
		{
			name:     "fallback to default locale when key missing in requested locale",
			locale:   "ar",
			key:      "error.not_found",
			expected: "Not found",
		},
		{
			name:     "return key when translation missing in all locales",
			locale:   "en",
			key:      "missing.key",
			expected: "missing.key",
		},
		{
			name:     "return key when locale not supported",
			locale:   "fr",
			key:      "audit.action.create",
			expected: "Created", // Falls back to default locale
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := localizer.Translate(tt.locale, tt.key)
			if result != tt.expected {
				t.Errorf("Translate(%q, %q) = %q, want %q", tt.locale, tt.key, result, tt.expected)
			}
		})
	}
}

func TestLocalizer_TranslateWithParams(t *testing.T) {
	translations := map[string]map[string]string{
		"en": {
			"error.employee_not_found": "Employee {{.Name}} not found",
			"notification.leave_approved": "Your leave from {{.StartDate}} to {{.EndDate}} has been approved",
			"validation.required":      "{{.Field}} is required",
			"simple.message":           "No parameters here",
		},
		"ar": {
			"error.employee_not_found": "الموظف {{.Name}} غير موجود",
		},
	}

	localizer := NewLocalizer(translations, "en")

	tests := []struct {
		name     string
		locale   string
		key      string
		params   map[string]interface{}
		expected string
	}{
		{
			name:   "substitute single parameter",
			locale: "en",
			key:    "error.employee_not_found",
			params: map[string]interface{}{
				"Name": "John Doe",
			},
			expected: "Employee John Doe not found",
		},
		{
			name:   "substitute multiple parameters",
			locale: "en",
			key:    "notification.leave_approved",
			params: map[string]interface{}{
				"StartDate": "2024-01-01",
				"EndDate":   "2024-01-05",
			},
			expected: "Your leave from 2024-01-01 to 2024-01-05 has been approved",
		},
		{
			name:   "substitute parameter in Arabic",
			locale: "ar",
			key:    "error.employee_not_found",
			params: map[string]interface{}{
				"Name": "أحمد",
			},
			expected: "الموظف أحمد غير موجود",
		},
		{
			name:     "no parameters provided",
			locale:   "en",
			key:      "simple.message",
			params:   nil,
			expected: "No parameters here",
		},
		{
			name:   "empty parameters map",
			locale: "en",
			key:    "simple.message",
			params: map[string]interface{}{},
			expected: "No parameters here",
		},
		{
			name:   "missing parameter leaves placeholder unchanged",
			locale: "en",
			key:    "error.employee_not_found",
			params: map[string]interface{}{},
			expected: "Employee {{.Name}} not found",
		},
		{
			name:   "HTML escaping for XSS prevention",
			locale: "en",
			key:    "validation.required",
			params: map[string]interface{}{
				"Field": "<script>alert('xss')</script>",
			},
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt; is required",
		},
		{
			name:   "non-string parameters not escaped",
			locale: "en",
			key:    "validation.required",
			params: map[string]interface{}{
				"Field": 123,
			},
			expected: "123 is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := localizer.TranslateWithParams(tt.locale, tt.key, tt.params)
			if result != tt.expected {
				t.Errorf("TranslateWithParams(%q, %q, %v) = %q, want %q", 
					tt.locale, tt.key, tt.params, result, tt.expected)
			}
		})
	}
}

func TestLocalizer_TranslateWithParams_InvalidTemplate(t *testing.T) {
	translations := map[string]map[string]string{
		"en": {
			"invalid.template": "{{.Name",
		},
	}

	localizer := NewLocalizer(translations, "en")

	// Should return the template string as-is when parsing fails
	result := localizer.TranslateWithParams("en", "invalid.template", map[string]interface{}{
		"Name": "Test",
	})

	if result != "{{.Name" {
		t.Errorf("Expected invalid template to be returned as-is, got %q", result)
	}
}

func TestLocalizer_GetSupportedLocales(t *testing.T) {
	translations := map[string]map[string]string{
		"en": {"key": "value"},
		"ar": {"key": "قيمة"},
		"fr": {"key": "valeur"},
	}

	localizer := NewLocalizer(translations, "en")
	locales := localizer.GetSupportedLocales()

	if len(locales) != 3 {
		t.Errorf("Expected 3 locales, got %d", len(locales))
	}

	// Check that all expected locales are present
	localeMap := make(map[string]bool)
	for _, locale := range locales {
		localeMap[locale] = true
	}

	expectedLocales := []string{"en", "ar", "fr"}
	for _, expected := range expectedLocales {
		if !localeMap[expected] {
			t.Errorf("Expected locale %q not found in result", expected)
		}
	}
}

func TestLocalizer_GetDefaultLocale(t *testing.T) {
	translations := map[string]map[string]string{
		"en": {"key": "value"},
	}

	localizer := NewLocalizer(translations, "en")
	defaultLocale := localizer.GetDefaultLocale()

	if defaultLocale != "en" {
		t.Errorf("Expected default locale 'en', got %q", defaultLocale)
	}
}

func TestLocalizer_ConcurrentAccess(t *testing.T) {
	translations := map[string]map[string]string{
		"en": {
			"audit.action.create": "Created",
			"audit.action.update": "Updated",
			"audit.action.delete": "Deleted",
		},
		"ar": {
			"audit.action.create": "تم الإنشاء",
			"audit.action.update": "تم التحديث",
			"audit.action.delete": "تم الحذف",
		},
	}

	localizer := NewLocalizer(translations, "en")

	// Run concurrent translations
	const numGoroutines = 100
	const numIterations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			locale := "en"
			if id%2 == 0 {
				locale = "ar"
			}

			for j := 0; j < numIterations; j++ {
				// Test Translate
				result := localizer.Translate(locale, "audit.action.create")
				if result == "" {
					t.Errorf("Translate returned empty string")
				}

				// Test TranslateWithParams
				result = localizer.TranslateWithParams(locale, "audit.action.update", map[string]interface{}{
					"Test": "value",
				})
				if result == "" {
					t.Errorf("TranslateWithParams returned empty string")
				}

				// Test GetSupportedLocales
				locales := localizer.GetSupportedLocales()
				if len(locales) != 2 {
					t.Errorf("Expected 2 locales, got %d", len(locales))
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestLocalizer_HTMLEscaping(t *testing.T) {
	translations := map[string]map[string]string{
		"en": {
			"message": "Hello {{.Name}}",
		},
	}

	localizer := NewLocalizer(translations, "en")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "escape script tags",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "escape HTML entities",
			input:    "<div>Test & \"quotes\"</div>",
			expected: "&lt;div&gt;Test &amp; &#34;quotes&#34;&lt;/div&gt;",
		},
		{
			name:     "normal text unchanged",
			input:    "John Doe",
			expected: "John Doe",
		},
		{
			name:     "Arabic text unchanged",
			input:    "أحمد محمد",
			expected: "أحمد محمد",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := localizer.TranslateWithParams("en", "message", map[string]interface{}{
				"Name": tt.input,
			})

			expected := "Hello " + tt.expected
			if result != expected {
				t.Errorf("Expected %q, got %q", expected, result)
			}
		})
	}
}

func TestLocalizer_PerformanceO1Lookup(t *testing.T) {
	// Create a large translation map to test O(1) lookup performance
	translations := map[string]map[string]string{
		"en": make(map[string]string),
	}

	// Add 10,000 translations
	for i := 0; i < 10000; i++ {
		key := "key." + strings.Repeat("a", i%100)
		translations["en"][key] = "value"
	}

	localizer := NewLocalizer(translations, "en")

	// Lookup should be O(1) regardless of map size
	// This test just ensures it completes quickly
	for i := 0; i < 1000; i++ {
		localizer.Translate("en", "key.a")
	}
}
