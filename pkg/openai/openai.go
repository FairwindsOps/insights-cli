package openai

import (
	"context"
	"strings"
)

type Client struct {
	apiKey string
	model  string
	debug  bool
}

const promptBoilerplate = `
Here is an example OPA policy. Can you write a brand new OPA policy example using the 'fairwinds' package?

<>
package fairwinds

labelblock[actionItem] {
    # List the keys of Kubernetes labels that will be blocked.
    blockedLabels := {"foo"}
    provided := {label | input.metadata.labels[label]}
    missing := blockedLabels - provided
    found := blockedLabels - missing
    count(found) > 0
    actionItem := {
        "title": "Bad label is present",
        "description": sprintf("Label %v is present", [found]),
        "severity": 0.1,
        "remediation": "Remove the label",
        "category": "Reliability"
    }
}
<>

The new policy you write should: '`

func NewClient(apiKey, model string, debug bool) *Client {
	return &Client{
		apiKey: apiKey,
		model:  model,
		debug:  debug,
	}
}

func (cs *Client) SendPrompt(ctx context.Context, prompt string) (string, error) {
	message := Message{Role: "user", Content: promptBoilerplate + prompt + "'"}
	request := Request{Model: cs.model, Messages: []Message{message}}
	content, err := sendRequest(ctx, cs.apiKey, request)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(content, "<>", ""), nil // removes the delimiters
}
