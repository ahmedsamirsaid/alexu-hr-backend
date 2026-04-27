package http

import "testing"

func TestHolidayVisibleToDepartmentUIDs_UnrestrictedAccess(t *testing.T) {
	if !holidayVisibleToDepartmentUIDs([]string{"dep-finance"}, nil) {
		t.Fatal("expected scoped holiday to be visible for unrestricted access")
	}
}

func TestHolidayVisibleToDepartmentUIDs_RestrictedNoMatch(t *testing.T) {
	if holidayVisibleToDepartmentUIDs([]string{"dep-finance"}, []string{"dep-hr"}) {
		t.Fatal("expected scoped holiday to be hidden when departments do not match")
	}
}

func TestHolidayVisibleToDepartmentUIDs_GlobalHolidayAlwaysVisible(t *testing.T) {
	if !holidayVisibleToDepartmentUIDs(nil, []string{"dep-hr"}) {
		t.Fatal("expected global holiday to remain visible")
	}
}
