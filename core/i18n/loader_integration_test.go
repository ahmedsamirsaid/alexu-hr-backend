package i18n

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoader_Integration demonstrates the loader working with realistic translation files
func TestLoader_Integration(t *testing.T) {
	// Create temporary directory structure mimicking real locales directory
	tmpDir := t.TempDir()

	// Create English locale directory
	enDir := filepath.Join(tmpDir, "en")
	if err := os.Mkdir(enDir, 0755); err != nil {
		t.Fatalf("Failed to create en directory: %v", err)
	}

	// Create Arabic locale directory
	arDir := filepath.Join(tmpDir, "ar")
	if err := os.Mkdir(arDir, 0755); err != nil {
		t.Fatalf("Failed to create ar directory: %v", err)
	}

	// Create English audit translations
	enAuditJSON := `{
		"audit": {
			"action": {
				"create": "Created",
				"update": "Updated",
				"delete": "Deleted",
				"approve": "Approved",
				"reject": "Rejected"
			},
			"entity": {
				"employee": "Employee",
				"leave_request": "Leave Request",
				"attendance_record": "Attendance Record"
			},
			"metadata": {
				"old_value": "Old Value",
				"new_value": "New Value",
				"reason": "Reason"
			}
		}
	}`
	if err := os.WriteFile(filepath.Join(enDir, "audit.json"), []byte(enAuditJSON), 0644); err != nil {
		t.Fatalf("Failed to create en/audit.json: %v", err)
	}

	// Create English error translations
	enErrorsJSON := `{
		"error": {
			"employee": {
				"not_found": "Employee not found",
				"already_exists": "Employee already exists"
			},
			"leave": {
				"insufficient_balance": "Insufficient leave balance",
				"invalid_date_range": "Invalid date range"
			}
		}
	}`
	if err := os.WriteFile(filepath.Join(enDir, "errors.json"), []byte(enErrorsJSON), 0644); err != nil {
		t.Fatalf("Failed to create en/errors.json: %v", err)
	}

	// Create Arabic audit translations
	arAuditJSON := `{
		"audit": {
			"action": {
				"create": "تم الإنشاء",
				"update": "تم التحديث",
				"delete": "تم الحذف",
				"approve": "تمت الموافقة",
				"reject": "تم الرفض"
			},
			"entity": {
				"employee": "موظف",
				"leave_request": "طلب إجازة",
				"attendance_record": "سجل الحضور"
			},
			"metadata": {
				"old_value": "القيمة القديمة",
				"new_value": "القيمة الجديدة",
				"reason": "السبب"
			}
		}
	}`
	if err := os.WriteFile(filepath.Join(arDir, "audit.json"), []byte(arAuditJSON), 0644); err != nil {
		t.Fatalf("Failed to create ar/audit.json: %v", err)
	}

	// Create Arabic error translations
	arErrorsJSON := `{
		"error": {
			"employee": {
				"not_found": "الموظف غير موجود",
				"already_exists": "الموظف موجود بالفعل"
			},
			"leave": {
				"insufficient_balance": "رصيد الإجازة غير كافٍ",
				"invalid_date_range": "نطاق التاريخ غير صالح"
			}
		}
	}`
	if err := os.WriteFile(filepath.Join(arDir, "errors.json"), []byte(arErrorsJSON), 0644); err != nil {
		t.Fatalf("Failed to create ar/errors.json: %v", err)
	}

	// Create loader and load all translations
	loader := NewLoader(tmpDir)
	allTranslations, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("Failed to load all translations: %v", err)
	}

	// Verify both locales were loaded
	if len(allTranslations) != 2 {
		t.Errorf("Expected 2 locales, got %d", len(allTranslations))
	}

	// Verify English translations
	enTranslations, exists := allTranslations["en"]
	if !exists {
		t.Fatal("English locale not loaded")
	}

	enTests := map[string]string{
		"audit.action.create":              "Created",
		"audit.action.approve":             "Approved",
		"audit.entity.employee":            "Employee",
		"audit.entity.leave_request":       "Leave Request",
		"audit.metadata.old_value":         "Old Value",
		"error.employee.not_found":         "Employee not found",
		"error.leave.insufficient_balance": "Insufficient leave balance",
	}

	for key, expected := range enTests {
		if got, exists := enTranslations[key]; !exists {
			t.Errorf("English translation missing for key: %s", key)
		} else if got != expected {
			t.Errorf("English translation for %s: expected %q, got %q", key, expected, got)
		}
	}

	// Verify Arabic translations
	arTranslations, exists := allTranslations["ar"]
	if !exists {
		t.Fatal("Arabic locale not loaded")
	}

	arTests := map[string]string{
		"audit.action.create":              "تم الإنشاء",
		"audit.action.approve":             "تمت الموافقة",
		"audit.entity.employee":            "موظف",
		"audit.entity.leave_request":       "طلب إجازة",
		"audit.metadata.old_value":         "القيمة القديمة",
		"error.employee.not_found":         "الموظف غير موجود",
		"error.leave.insufficient_balance": "رصيد الإجازة غير كافٍ",
	}

	for key, expected := range arTests {
		if got, exists := arTranslations[key]; !exists {
			t.Errorf("Arabic translation missing for key: %s", key)
		} else if got != expected {
			t.Errorf("Arabic translation for %s: expected %q, got %q", key, expected, got)
		}
	}

	// Verify translation counts
	t.Logf("Loaded %d English translations", len(enTranslations))
	t.Logf("Loaded %d Arabic translations", len(arTranslations))

	// Both locales should have the same number of keys
	if len(enTranslations) != len(arTranslations) {
		t.Errorf("Translation count mismatch: English has %d, Arabic has %d", len(enTranslations), len(arTranslations))
	}
}

// TestLoader_LoadLocale_RealWorldScenario tests loading a single locale
func TestLoader_LoadLocale_RealWorldScenario(t *testing.T) {
	tmpDir := t.TempDir()
	enDir := filepath.Join(tmpDir, "en")
	if err := os.Mkdir(enDir, 0755); err != nil {
		t.Fatalf("Failed to create en directory: %v", err)
	}

	// Create multiple translation files
	files := map[string]string{
		"audit.json": `{
			"audit": {
				"action": {
					"create": "Created",
					"update": "Updated"
				}
			}
		}`,
		"errors.json": `{
			"error": {
				"validation": {
					"required_field": "{{.Field}} is required"
				}
			}
		}`,
		"notifications.json": `{
			"notification": {
				"leave_approved": {
					"title": "Leave Request Approved",
					"body": "Your leave request has been approved"
				}
			}
		}`,
	}

	for filename, content := range files {
		if err := os.WriteFile(filepath.Join(enDir, filename), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", filename, err)
		}
	}

	loader := NewLoader(tmpDir)
	translations, err := loader.LoadLocale("en")
	if err != nil {
		t.Fatalf("Failed to load locale: %v", err)
	}

	// Verify all translations from all files were loaded
	expectedKeys := []string{
		"audit.action.create",
		"audit.action.update",
		"error.validation.required_field",
		"notification.leave_approved.title",
		"notification.leave_approved.body",
	}

	for _, key := range expectedKeys {
		if _, exists := translations[key]; !exists {
			t.Errorf("Expected key %s not found in translations", key)
		}
	}

	// Verify parameterized template is preserved
	if got := translations["error.validation.required_field"]; got != "{{.Field}} is required" {
		t.Errorf("Parameterized template not preserved: got %q", got)
	}
}
