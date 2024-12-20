package internal

import "testing"

func TestIsValidURL(t *testing.T) {
	testCases := []struct {
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
