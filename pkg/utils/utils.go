package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:length], nil
}

// ToJSON converts a value to JSON string
func ToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// Contains checks if a string slice contains a value
func Contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// ContainsIgnoreCase checks if a string slice contains a value (case-insensitive)
func ContainsIgnoreCase(slice []string, value string) bool {
	value = strings.ToLower(value)
	for _, item := range slice {
		if strings.ToLower(item) == value {
			return true
		}
	}
	return false
}

// TruncateString truncates a string to specified length
func TruncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// SafeError returns a safe error message (hides sensitive info in production)
func SafeError(err error, devMode bool) string {
	if devMode {
		return err.Error()
	}
	return "An error occurred"
}

// PrettyPrint prints a value as pretty JSON
func PrettyPrint(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(b)
}
