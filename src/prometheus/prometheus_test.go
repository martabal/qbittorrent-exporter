package prom

import (
	app "qbit-exp/app"
	"testing"
)

func TestMain(t *testing.T) {
	app.SetVar(0, false, "", "http://localhost:8080", "admin", "adminadmin", 30, false)
	result := app.GetPasswordMasked()

	if !isValidMaskedPassword(result) {
		t.Errorf("Invalid masked password. Expected only asterisks, got: %s", result)
	}
}

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
			result := isValidURL(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %s to be %v, but got %v", tc.input, tc.expected, result)
			}
		})
	}
}

func isValidMaskedPassword(password string) bool {
	for _, char := range password {
		if char != '*' {
			return false
		}
	}
	return true
}
