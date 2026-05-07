package ports

import "context"

// I18nService defines the interface for internationalization.
// This allows the core layer to depend on an abstraction.
type I18nService interface {
	// T translates a key to the locale in the context.
	// Returns the key itself if translation is missing.
	T(ctx context.Context, key string) string

	// TWithParams translates a key with parameter substitution.
	// Parameters are substituted using Go template syntax: {{.ParamName}}
	TWithParams(ctx context.Context, key string, params map[string]interface{}) string

	// TLocale translates a key to a specific locale, bypassing context.
	TLocale(locale, key string) string

	// TLocaleWithParams translates with params to a specific locale.
	TLocaleWithParams(locale, key string, params map[string]interface{}) string

	// GetSupportedLocales returns list of available locales.
	GetSupportedLocales() []string

	// GetDefaultLocale returns the fallback locale.
	GetDefaultLocale() string
}
