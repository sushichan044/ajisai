package utils_test

import (
	"reflect"
	"testing"

	"github.com/sushichan044/ajisai/utils"
)

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name     string
		source   []int
		values   []int
		expected bool
	}{
		{
			name:     "Empty source slice",
			source:   []int{},
			values:   []int{1, 2, 3},
			expected: false,
		},
		{
			name:     "Empty values slice",
			source:   []int{1, 2, 3},
			values:   []int{},
			expected: false,
		},
		{
			name:     "Has common value",
			source:   []int{1, 2, 3},
			values:   []int{3, 4, 5},
			expected: true,
		},
		{
			name:     "No common value",
			source:   []int{1, 2, 3},
			values:   []int{4, 5, 6},
			expected: false,
		},
		{
			name:     "Multiple common values",
			source:   []int{1, 2, 3, 4},
			values:   []int{2, 4, 6},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ContainsAny(tt.source, tt.values)
			if result != tt.expected {
				t.Errorf("ContainsAny() = %v, want %v", result, tt.expected)
			}
		})
	}

	stringTests := []struct {
		name     string
		source   []string
		values   []string
		expected bool
	}{
		{
			name:     "Has common value",
			source:   []string{"apple", "banana", "cherry"},
			values:   []string{"cherry", "date", "fig"},
			expected: true,
		},
		{
			name:     "No common value",
			source:   []string{"apple", "banana", "cherry"},
			values:   []string{"date", "fig", "grape"},
			expected: false,
		},
	}

	for _, tt := range stringTests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ContainsAny(tt.source, tt.values)
			if result != tt.expected {
				t.Errorf("ContainsAny() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRemoveZeroValues(t *testing.T) {
	tests := []struct {
		name     string
		source   []int
		expected []int
	}{
		{
			name:     "Integer slice with zero values",
			source:   []int{0, 1, 0, 2, 0, 3},
			expected: []int{1, 2, 3},
		},
		{
			name:     "Integer slice without zero values",
			source:   []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
		{
			name:     "All zero values integer slice",
			source:   []int{0, 0, 0},
			expected: []int{},
		},
		{
			name:     "Empty integer slice",
			source:   []int{},
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.RemoveZeroValues(tt.source)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("RemoveZeroValues() = %v, want %v", result, tt.expected)
			}
		})
	}

	stringTests := []struct {
		name     string
		source   []string
		expected []string
	}{
		{
			name:     "Empty strings in string slice",
			source:   []string{"", "hello", "", "world", ""},
			expected: []string{"hello", "world"},
		},
		{
			name:     "No empty strings in string slice",
			source:   []string{"hello", "world"},
			expected: []string{"hello", "world"},
		},
	}

	for _, tt := range stringTests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.RemoveZeroValues(tt.source)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("RemoveZeroValues() = %v, want %v", result, tt.expected)
			}
		})
	}

	type testStruct struct {
		Value int
		Name  string
	}

	structTests := []struct {
		name     string
		source   []testStruct
		expected []testStruct
	}{
		{
			name: "Slice containing zero values struct",
			source: []testStruct{
				{Value: 0, Name: ""},
				{Value: 1, Name: "one"},
				{Value: 0, Name: ""},
				{Value: 2, Name: "two"},
			},
			expected: []testStruct{
				{Value: 1, Name: "one"},
				{Value: 2, Name: "two"},
			},
		},
	}

	for _, tt := range structTests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.RemoveZeroValues(tt.source)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("RemoveZeroValues() = %v, want %v", result, tt.expected)
			}
		})
	}
}
