package crawler

import (
	"reflect"
	"testing"
)

func TestParseWeeks(t *testing.T) {
	tests := []struct {
		input    string
		expected []int64
	}{
		{"1-6周", []int64{1, 2, 3, 4, 5, 6}},
		{"1-6周,8-9周", []int64{1, 2, 3, 4, 5, 6, 8, 9}},
		{"1-6周(单)", []int64{1, 3, 5}},
		{"1-6周(双)", []int64{2, 4, 6}},
		{"8周,11-15周(单)", []int64{8, 11, 13, 15}},
	}

	for _, tt := range tests {
		result, err := ParseWeeks(tt.input)
		if err != nil {
			t.Fatalf("parseWeeks(%s) returned an error: %v", tt.input, err)
		}
		if !reflect.DeepEqual(result, tt.expected) {
			t.Errorf("parseWeeks(%s) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}
