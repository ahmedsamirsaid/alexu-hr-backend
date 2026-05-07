package i18n

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoader_LoadJSONFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		jsonContent string
		expected    map[string]string
		expectError bool
	}{
		{
			name: "simple flat structure",
			jsonContent: `{
				"hello": "Hello",
				"goodbye": "Goodbye"
			}`,
			expected: map[string]string{
				"hello":   "Hello",
				"goodbye": "Goodbye",
			},
			expectError: false,
		},
		{
			name: "nested structure with dot notation",
			jsonContent: `{
				"audit": {
					"action": {
						"create": "Created",
						"update": "Updated",
						"delete": "Deleted"
					},
					"entity": {
						"employee": "Employee",
						"leave_request": "Leave Request"
					}
				}
			}`,
			expected: map[string]string{
				"audit.action.create":          "Created",
				"audit.action.update":          "Updated",
				"audit.action.delete":          "Deleted",
				"audit.entity.employee":        "Employee",
				"audit.entity.leave_request":   "Leave Request",
			},
			expectError: false,
		},
		{
			name: "deeply nested structure",
			jsonContent: `{
				"level1": {
					"level2": {
						"level3": {
							"level4": "Deep Value"
						}
					}
				}
			}`,
			expected: map[string]string{
				"level1.level2.level3.level4": "Deep Value",
			},
			expectError: false,
		},
		{
			name: "mixed flat and nested",
			jsonContent: `{
				"simple": "Simple Value",
				"nested": {
					"key": "Nested Value"
				}
			}`,
			expected: map[string]string{
				"simple":     "Simple Value",
				"nested.key": "Nested Value",
			},
			expectError: false,
		},
		{
			name:        "malformed JSON",
			jsonContent: `{"invalid": "json"`,
			expected:    nil,
			expectError: true,
		},
		{
			name: "empty JSON object",
			jsonContent: `{}`,
			expected:    map[string]string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tmpDir, "test.json")
			err := os.WriteFile(testFile, []byte(tt.jsonContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
			defer os.Remove(testFile)

			// Load the file
			loader := NewLoader(tmpDir)
			result, err := loader.loadJSONFile(testFile)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check results
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d translations, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if gotValue, exists := result[key]; !exists {
					t.Errorf("Missing key: %s", key)
				} else if gotValue != expectedValue {
					t.Errorf("Key %s: expected %q, got %q", key, expectedValue, gotValue)
				}
			}
		})
	}
}

func TestLoader_LoadLocale(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	enDir := filepath.Join(tmpDir, "en")
	err := os.Mkdir(enDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create en directory: %v", err)
	}

	// Create test translation files
	auditJSON := `{
		"audit": {
			"action": {
				"create": "Created",
				"update": "Updated"
			}
		}
	}`
	err = os.WriteFile(filepath.Join(enDir, "audit.json"), []byte(auditJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create audit.json: %v", err)
	}

	errorsJSON := `{
		"error": {
			"employee": {
				"not_found": "Employee not found"
			}
		}
	}`
	err = os.WriteFile(filepath.Join(enDir, "errors.json"), []byte(errorsJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create errors.json: %v", err)
	}

	// Create a non-JSON file that should be ignored
	err = os.WriteFile(filepath.Join(enDir, "readme.txt"), []byte("This should be ignored"), 0644)
	if err != nil {
		t.Fatalf("Failed to create readme.txt: %v", err)
	}

	loader := NewLoader(tmpDir)
	result, err := loader.LoadLocale("en")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := map[string]string{
		"audit.action.create":       "Created",
		"audit.action.update":       "Updated",
		"error.employee.not_found":  "Employee not found",
	}

	if len(result) != len(expected) {
		t.Errorf("Expected %d translations, got %d", len(expected), len(result))
	}

	for key, expectedValue := range expected {
		if gotValue, exists := result[key]; !exists {
			t.Errorf("Missing key: %s", key)
		} else if gotValue != expectedValue {
			t.Errorf("Key %s: expected %q, got %q", key, expectedValue, gotValue)
		}
	}
}

func TestLoader_LoadLocale_NonExistentLocale(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoader(tmpDir)

	_, err := loader.LoadLocale("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent locale, got none")
	}
}

func TestLoader_LoadLocale_MalformedJSON(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	enDir := filepath.Join(tmpDir, "en")
	err := os.Mkdir(enDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create en directory: %v", err)
	}

	// Create valid JSON file
	validJSON := `{"valid": "Valid Translation"}`
	err = os.WriteFile(filepath.Join(enDir, "valid.json"), []byte(validJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create valid.json: %v", err)
	}

	// Create malformed JSON file
	malformedJSON := `{"invalid": "json"`
	err = os.WriteFile(filepath.Join(enDir, "malformed.json"), []byte(malformedJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create malformed.json: %v", err)
	}

	loader := NewLoader(tmpDir)
	result, err := loader.LoadLocale("en")

	// Should not return error, but should log and continue with valid files
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have loaded the valid file
	if result["valid"] != "Valid Translation" {
		t.Error("Valid translation should have been loaded despite malformed file")
	}

	// Should not have loaded the malformed file
	if _, exists := result["invalid"]; exists {
		t.Error("Malformed file should not have been loaded")
	}
}

func TestLoader_LoadAll(t *testing.T) {
	// Create temporary directory structure with multiple locales
	tmpDir := t.TempDir()

	// Create English locale
	enDir := filepath.Join(tmpDir, "en")
	err := os.Mkdir(enDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create en directory: %v", err)
	}
	enJSON := `{"greeting": "Hello"}`
	err = os.WriteFile(filepath.Join(enDir, "common.json"), []byte(enJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create en/common.json: %v", err)
	}

	// Create Arabic locale
	arDir := filepath.Join(tmpDir, "ar")
	err = os.Mkdir(arDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create ar directory: %v", err)
	}
	arJSON := `{"greeting": "مرحبا"}`
	err = os.WriteFile(filepath.Join(arDir, "common.json"), []byte(arJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create ar/common.json: %v", err)
	}

	// Create a file in the root that should be ignored
	err = os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("Ignore me"), 0644)
	if err != nil {
		t.Fatalf("Failed to create readme.txt: %v", err)
	}

	loader := NewLoader(tmpDir)
	result, err := loader.LoadAll()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have loaded both locales
	if len(result) != 2 {
		t.Errorf("Expected 2 locales, got %d", len(result))
	}

	// Check English translations
	if enTranslations, exists := result["en"]; !exists {
		t.Error("English locale not loaded")
	} else if enTranslations["greeting"] != "Hello" {
		t.Errorf("English greeting: expected %q, got %q", "Hello", enTranslations["greeting"])
	}

	// Check Arabic translations
	if arTranslations, exists := result["ar"]; !exists {
		t.Error("Arabic locale not loaded")
	} else if arTranslations["greeting"] != "مرحبا" {
		t.Errorf("Arabic greeting: expected %q, got %q", "مرحبا", arTranslations["greeting"])
	}
}

func TestLoader_LoadAll_NonExistentDirectory(t *testing.T) {
	loader := NewLoader("/nonexistent/path")
	_, err := loader.LoadAll()

	if err == nil {
		t.Error("Expected error for non-existent directory, got none")
	}
}

func TestLoader_LoadAll_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewLoader(tmpDir)

	_, err := loader.LoadAll()
	if err == nil {
		t.Error("Expected error for empty directory, got none")
	}
}

func TestLoader_LoadAll_OnlyFiles(t *testing.T) {
	// Create directory with only files, no locale subdirectories
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "file.json"), []byte(`{"key": "value"}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	loader := NewLoader(tmpDir)
	_, err = loader.LoadAll()

	if err == nil {
		t.Error("Expected error when no locale directories exist, got none")
	}
}

func TestLoader_DuplicateKeys(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	enDir := filepath.Join(tmpDir, "en")
	err := os.Mkdir(enDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create en directory: %v", err)
	}

	// Create two files with duplicate keys
	file1JSON := `{"duplicate": "First Value"}`
	err = os.WriteFile(filepath.Join(enDir, "file1.json"), []byte(file1JSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create file1.json: %v", err)
	}

	file2JSON := `{"duplicate": "Second Value"}`
	err = os.WriteFile(filepath.Join(enDir, "file2.json"), []byte(file2JSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create file2.json: %v", err)
	}

	loader := NewLoader(tmpDir)
	result, err := loader.LoadLocale("en")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have loaded one of the values (last one wins)
	if _, exists := result["duplicate"]; !exists {
		t.Error("Duplicate key should have been loaded")
	}
	// The exact value depends on file read order, but it should be one of them
	value := result["duplicate"]
	if value != "First Value" && value != "Second Value" {
		t.Errorf("Unexpected value for duplicate key: %q", value)
	}
}

func TestLoader_InvalidValueTypes(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()
	enDir := filepath.Join(tmpDir, "en")
	err := os.Mkdir(enDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create en directory: %v", err)
	}

	// Create JSON with invalid value types (numbers, arrays, etc.)
	invalidJSON := `{
		"valid": "Valid String",
		"number": 123,
		"boolean": true,
		"array": ["item1", "item2"],
		"null": null,
		"nested": {
			"valid": "Nested Valid",
			"invalid": 456
		}
	}`
	err = os.WriteFile(filepath.Join(enDir, "invalid.json"), []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid.json: %v", err)
	}

	loader := NewLoader(tmpDir)
	result, err := loader.LoadLocale("en")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have loaded only valid string values
	if result["valid"] != "Valid String" {
		t.Error("Valid string should have been loaded")
	}

	if result["nested.valid"] != "Nested Valid" {
		t.Error("Nested valid string should have been loaded")
	}

	// Invalid types should be skipped (logged as warnings)
	if _, exists := result["number"]; exists {
		t.Error("Number value should not have been loaded")
	}

	if _, exists := result["boolean"]; exists {
		t.Error("Boolean value should not have been loaded")
	}

	if _, exists := result["array"]; exists {
		t.Error("Array value should not have been loaded")
	}
}

func TestNewLoader(t *testing.T) {
	basePath := "/test/path"
	loader := NewLoader(basePath)

	if loader == nil {
		t.Fatal("NewLoader returned nil")
	}

	if loader.basePath != basePath {
		t.Errorf("Expected basePath %q, got %q", basePath, loader.basePath)
	}
}
