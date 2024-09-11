package utils

import (
	"fmt"

	"github.com/imroc/req"
)

func GetHeaders(version, token string) req.Header {
	return req.Header{
		"X-Fairwinds-CLI-Version": version,
		"Authorization":           fmt.Sprintf("Bearer %s", token),
		"Accept":                  "application/json",
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
