package openai

import "fmt"

type Operation struct {
	Method       string
	Filenames    []string
	Content      []string
	OriginalText string
}

func parseResponse(content string) *Operation {
	// TODO: Verify response format and properly format it
	fmt.Println(content)
	return nil
}
