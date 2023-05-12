package openai

type Operation struct {
	Method       string
	Filenames    []string
	Content      []string
	OriginalText string
}
