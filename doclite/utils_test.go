package doclite

import (
	"testing"
)

// TestToMapZeroValuedFloatsAndUnsignedInts verifies that zero-valued float32,
// float64, and unsigned integer fields in a struct filter are included in the
// map produced by toMap. This is a regression test for the bug where those
// types were incorrectly skipped when their value was the zero value.
func TestToMapZeroValuedFloatsAndUnsignedInts(t *testing.T) {
	type filterStruct struct {
		Name    string
		Float32 float32
		Float64 float64
		Uint    uint
		Uint8   uint8
		Uint16  uint16
		Uint32  uint32
		Uint64  uint64
		Int     int
		Int32   int32
	}

	f := &filterStruct{
		Name:    "test",
		Float32: 0,
		Float64: 0,
		Uint:    0,
		Uint8:   0,
		Uint16:  0,
		Uint32:  0,
		Uint64:  0,
		Int:     0,
		Int32:   0,
	}

	m := toMap(f)

	// All zero-valued numeric fields should be present
	expectedFields := []string{"Float32", "Float64", "Uint", "Uint8", "Uint16", "Uint32", "Uint64", "Int", "Int32", "Name"}
	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("expected field %q to be present in map, but it was missing", field)
		}
	}

	// Verify the values are correct
	if m["Float32"].(float32) != 0 {
		t.Errorf("expected Float32=0, got %v", m["Float32"])
	}
	if m["Float64"].(float64) != 0 {
		t.Errorf("expected Float64=0, got %v", m["Float64"])
	}
	if m["Uint"].(uint) != 0 {
		t.Errorf("expected Uint=0, got %v", m["Uint"])
	}
	if m["Uint8"].(uint8) != 0 {
		t.Errorf("expected Uint8=0, got %v", m["Uint8"])
	}
	if m["Uint16"].(uint16) != 0 {
		t.Errorf("expected Uint16=0, got %v", m["Uint16"])
	}
	if m["Uint32"].(uint32) != 0 {
		t.Errorf("expected Uint32=0, got %v", m["Uint32"])
	}
	if m["Uint64"].(uint64) != 0 {
		t.Errorf("expected Uint64=0, got %v", m["Uint64"])
	}
}

// TestToMapNonZeroValuesStillWork verifies that non-zero values are included
// in the map (regression check to ensure the fix didn't break normal behavior).
func TestToMapNonZeroValuesStillWork(t *testing.T) {
	type filterStruct struct {
		Name    string
		Float32 float32
		Float64 float64
		Uint    uint
		Uint32  uint32
		Int     int
	}

	f := &filterStruct{
		Name:    "hello",
		Float32: 3.14,
		Float64: 2.718,
		Uint:    42,
		Uint32:  100,
		Int:     -5,
	}

	m := toMap(f)

	if m["Name"] != "hello" {
		t.Errorf("expected Name=\"hello\", got %v", m["Name"])
	}
	if m["Float32"].(float32) != 3.14 {
		t.Errorf("expected Float32=3.14, got %v", m["Float32"])
	}
	if m["Float64"].(float64) != 2.718 {
		t.Errorf("expected Float64=2.718, got %v", m["Float64"])
	}
	if m["Uint"].(uint) != 42 {
		t.Errorf("expected Uint=42, got %v", m["Uint"])
	}
	if m["Uint32"].(uint32) != 100 {
		t.Errorf("expected Uint32=100, got %v", m["Uint32"])
	}
	if m["Int"].(int) != -5 {
		t.Errorf("expected Int=-5, got %v", m["Int"])
	}
}

// TestToMapZeroValuedStringExcluded verifies that zero-valued string fields
// are still excluded (existing behavior should not change for non-numeric types).
func TestToMapZeroValuedStringExcluded(t *testing.T) {
	type filterStruct struct {
		Name  string
		Count int
	}

	f := &filterStruct{
		Name:  "",
		Count: 0,
	}

	m := toMap(f)

	if _, ok := m["Name"]; ok {
		t.Errorf("expected zero-valued string field to be excluded, but it was present")
	}
	if _, ok := m["Count"]; !ok {
		t.Errorf("expected zero-valued int field to be present, but it was missing")
	}
}
