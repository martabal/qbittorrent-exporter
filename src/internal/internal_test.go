package internal

import "testing"

func TestIsValidURL(t *testing.T) {
	testCases := [...]struct {
		input    string
		expected bool
	}{
		{"https://www.example.com", true},
		{"http://localhost:8080", true},
		{"ftp://ftp.example.com", true},
		{"not_a_url", false},
		{"www.example.com", false},
		{"file:///path/to/file", false},
		{"", false},
		{"http://invalid..url", true},
		{"https://[::1]", true},
		{"http://user:pass@www.example.com", true},
		{"https://www.example.com:8080/path", true},
		{"https://www.example.com?query=value", true},
		{"https://www.example.com#fragment", true},
		{"http://www.ex ample.com", false},
		{"http://www.example.com:10000", true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := IsValidURL(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %s to be %v, but got %v", tc.input, tc.expected, result)
			}
		})
	}
}

func TestIsValidHttpsURL(t *testing.T) {
	testCases := [...]struct {
		input    string
		expected bool
	}{
		{"https://www.example.com", true},
		{"http://localhost:8080", false},
		{"ftp://ftp.example.com", false},
		{"not_a_url", false},
		{"www.example.com", false},
		{"file:///path/to/file", false},
		{"", false},
		{"http://invalid..url", false},
		{"https://[::1]", true},
		{"http://user:pass@www.example.com", false},
		{"https://www.example.com:8080/path", true},
		{"https://www.example.com?query=value", true},
		{"https://www.example.com#fragment", true},
		{"http://www.ex ample.com", false},
		{"http://www.example.com:10000", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := IsValidHttpsURL(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %s to be %v, but got %v", tc.input, tc.expected, result)
			}
		})
	}
}

func TestEnsureLeadingSlash(t *testing.T) {
	tests := [...]struct {
		name           string
		input          *string
		expectedOutput string
		expectPanic    bool
	}{
		{
			name:           "already has leading slash",
			input:          strPtr("/example"),
			expectedOutput: "/example",
		},
		{
			name:           "missing leading slash",
			input:          strPtr("example"),
			expectedOutput: "/example",
		},
		{
			name:           "empty string",
			input:          strPtr(""),
			expectedOutput: "/",
		},
		{
			name:        "Nil input",
			input:       nil,
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic but did not panic")
					}
				}()
				EnsureLeadingSlash(tt.input)
			} else {
				EnsureLeadingSlash(tt.input)
				if *tt.input != tt.expectedOutput {
					t.Errorf("Expected %q, got %q", tt.expectedOutput, *tt.input)
				}
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func TestCompareSemVer(t *testing.T) {
	tests := [...]struct {
		v1, v2   string
		expected int
	}{
		{"1.2.3", "1.2.4", -1},  // v1 < v2
		{"1.2.10", "1.2.3", 1},  // v1 > v2
		{"1.2.0", "1.2.0", 0},   // v1 == v2
		{"1.2", "1.2.0", 0},     // v1 == v2 (implicit zero-padding)
		{"1.10.1", "1.9.9", 1},  // v1 > v2
		{"2.0.0", "1.9.9", 1},   // Major version difference
		{"1.0.0", "1.0.0", 0},   // Identical versions
		{"1.0", "1.0.1", -1},    // Missing patch version (implicit zero-padding)
		{"1.2.3.4", "1.2.3", 1}, // Extra segment in v1
		{"1.2", "1.2.1", -1},    // Minor version difference (implicit zero-padding)
		{"1.2.0", "1.2.1", -1},  // Minor version difference
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_"+tt.v2, func(t *testing.T) {
			result := CompareSemVer(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("compareSemVer(%q, %q) = %d; want %d", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}
