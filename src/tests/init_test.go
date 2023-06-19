package main

import (
	"qbit-exp/src/models"
	"testing"
)

func TestMain(t *testing.T) {
	models.SetQbit("http://localhost:8080", "admin", "adminadmin")
	result := models.Getpasswordmasked()

	if !isValidMaskedPassword(result) {
		t.Errorf("Invalid masked password. Expected only asterisks, got: %s", result)
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
