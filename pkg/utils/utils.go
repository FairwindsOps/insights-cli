package utils

func GetHeaders(inputContentType string) map[string]string {
	contentType := "application/json"
	if inputContentType != "" {
		contentType = inputContentType
	}
	return map[string]string{
		"Accept":       "application/json",
		"Content-Type": contentType,
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
