package i18n

import "context"

// Service provides translation capabilities across the application.
type Service interface {
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

// serviceImpl is the concrete implementation of the Service interface.
type serviceImpl struct {
	localizer        *Localizer
	supportedLocales []string
	defaultLocale    string
}

// NewService creates a new I18n service with the given translations and configuration.
// translations is a map of locale -> key -> value.
// defaultLocale is the fallback locale when translation is missing.
// supportedLocales is the list of locales that the service supports.
func NewService(translations map[string]map[string]string, defaultLocale string, supportedLocales []string) Service {
	localizer := NewLocalizer(translations, defaultLocale)
	return &serviceImpl{
		localizer:        localizer,
		supportedLocales: supportedLocales,
		defaultLocale:    defaultLocale,
	}
}

// T translates a key to the locale in the context.
// Returns the key itself if translation is missing.
func (s *serviceImpl) T(ctx context.Context, key string) string {
	locale := LocaleFromContext(ctx)
	return s.localizer.Translate(locale, key)
}

// TWithParams translates a key with parameter substitution.
// Parameters are substituted using Go template syntax: {{.ParamName}}
func (s *serviceImpl) TWithParams(ctx context.Context, key string, params map[string]interface{}) string {
	locale := LocaleFromContext(ctx)
	return s.localizer.TranslateWithParams(locale, key, params)
}

// TLocale translates a key to a specific locale, bypassing context.
func (s *serviceImpl) TLocale(locale, key string) string {
	return s.localizer.Translate(locale, key)
}

// TLocaleWithParams translates with params to a specific locale.
func (s *serviceImpl) TLocaleWithParams(locale, key string, params map[string]interface{}) string {
	return s.localizer.TranslateWithParams(locale, key, params)
}

// GetSupportedLocales returns list of available locales.
func (s *serviceImpl) GetSupportedLocales() []string {
	return s.supportedLocales
}

// GetDefaultLocale returns the fallback locale.
func (s *serviceImpl) GetDefaultLocale() string {
	return s.defaultLocale
}
