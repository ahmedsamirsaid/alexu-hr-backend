package i18n

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Loader loads translation templates from the filesystem.
type Loader struct {
	basePath string
}

// NewLoader creates a new Loader with the specified base path.
func NewLoader(basePath string) *Loader {
	return &Loader{
		basePath: basePath,
	}
}

// LoadAll loads all translation files from the locales directory.
// Returns a map of locale -> (key -> value).
// Scans the base path for locale directories and loads all JSON files within each.
func (l *Loader) LoadAll() (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)

	// Check if base path exists
	if _, err := os.Stat(l.basePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("locales directory does not exist: %s", l.basePath)
	}

	// Read locale directories
	entries, err := os.ReadDir(l.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read locales directory: %w", err)
	}

	// Load translations for each locale directory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		locale := entry.Name()
		translations, err := l.LoadLocale(locale)
		if err != nil {
			// Log error but continue with other locales
			slog.Error("i18n.loader.load_locale_failed",
				"locale", locale,
				"error", err,
			)
			continue
		}

		result[locale] = translations
		slog.Info("i18n.loader.locale_loaded",
			"locale", locale,
			"translation_count", len(translations),
		)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no valid locale directories found in %s", l.basePath)
	}

	return result, nil
}

// LoadLocale loads translation files for a specific locale.
// Returns a map of flattened keys to translation values.
func (l *Loader) LoadLocale(locale string) (map[string]string, error) {
	localePath := filepath.Join(l.basePath, locale)

	// Check if locale directory exists
	if _, err := os.Stat(localePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("locale directory does not exist: %s", localePath)
	}

	result := make(map[string]string)

	// Read all JSON files in the locale directory
	entries, err := os.ReadDir(localePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read locale directory %s: %w", localePath, err)
	}

	// Load each JSON file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .json files
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(localePath, entry.Name())
		translations, err := l.loadJSONFile(filePath)
		if err != nil {
			// Log error with detailed information but continue with other files
			slog.Error("i18n.loader.file_error",
				"file", filePath,
				"locale", locale,
				"error", err,
			)
			continue
		}

		// Merge translations into result
		for key, value := range translations {
			if existingValue, exists := result[key]; exists {
				// Log warning for duplicate keys
				slog.Warn("i18n.loader.duplicate_key",
					"locale", locale,
					"key", key,
					"old_value", existingValue,
					"new_value", value,
					"file", filePath,
				)
			}
			result[key] = value
		}
	}

	return result, nil
}

// loadJSONFile loads and parses a single JSON translation file.
// Returns a map of flattened dot-notation keys to translation values.
func (l *Loader) loadJSONFile(filePath string) (map[string]string, error) {
	// Read file contents
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON into nested structure
	var nested map[string]interface{}
	if err := json.Unmarshal(data, &nested); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Flatten nested structure into dot-notation keys
	result := make(map[string]string)
	l.flattenJSON(nested, "", result)

	return result, nil
}

// flattenJSON recursively flattens a nested JSON structure into dot-notation keys.
// Example: {"audit": {"action": {"create": "Created"}}} becomes "audit.action.create" -> "Created"
func (l *Loader) flattenJSON(data map[string]interface{}, prefix string, result map[string]string) {
	for key, value := range data {
		// Build the full key with dot notation
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			// Leaf node - store the translation
			result[fullKey] = v
		case map[string]interface{}:
			// Nested object - recurse
			l.flattenJSON(v, fullKey, result)
		default:
			// Invalid value type - log warning
			slog.Warn("i18n.loader.invalid_value_type",
				"key", fullKey,
				"type", fmt.Sprintf("%T", v),
			)
		}
	}
}
