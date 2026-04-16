package db

import (
	"testing"
	"time"
)

func TestMakeInclusiveEndDate_MidnightExpandsToEndOfDay(t *testing.T) {
	input := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)
	got := makeInclusiveEndDate(input)
	want := time.Date(2026, 4, 30, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)

	if !got.Equal(want) {
		t.Fatalf("makeInclusiveEndDate(%s) = %s, want %s", input, got, want)
	}
}

func TestMakeInclusiveEndDate_NonMidnightStaysUnchanged(t *testing.T) {
	input := time.Date(2026, 4, 30, 13, 45, 30, 500, time.UTC)
	got := makeInclusiveEndDate(input)

	if !got.Equal(input) {
		t.Fatalf("makeInclusiveEndDate(%s) = %s, want unchanged", input, got)
	}
}
