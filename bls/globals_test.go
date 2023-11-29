package bls

import (
	"testing"
)

func TestIsPowerOfTwo(t *testing.T) {
	testCases := []struct {
		name     string
		input    uint64
		expected bool
	}{
		{
			name:     "0 is not a power of 2",
			input:    0,
			expected: false,
		},
		{
			name:     "2^0 = 1",
			input:    1,
			expected: true,
		},
		{
			name:     "2^1 = 2",
			input:    2,
			expected: true,
		},
		{
			name:     "3 is not a power of 2",
			input:    3,
			expected: false,
		},
		{
			name:     "2^2 = 4",
			input:    4,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsPowerOfTwo(tc.input)
			if result != tc.expected {
				t.Fatalf("IsPowerOfTwo(%d) = %v; want %v", tc.input, result, tc.expected)
			}
		})
	}
}
