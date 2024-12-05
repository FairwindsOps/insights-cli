package utils

import (
	"fmt"
)

func GetHeaders(version, token, inputContentType string) map[string]string {
	contentType := "application/json"
	if inputContentType != "" {
		contentType = inputContentType
	}
	return map[string]string{
		"X-Fairwinds-CLI-Version": version,
		"Authorization":           fmt.Sprintf("Bearer %s", token),
		"Accept":                  "application/json",
		"Content-Type":            contentType,
	}
}

func IsSuccessful(statusCode int) bool { return statusCode >= 200 && statusCode < 400 }

func InvertBoolPointer(b *bool) *bool {
	if b == nil {
		return nil
	}
	r := !*b
	return &r
}
