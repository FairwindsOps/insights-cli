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
Can you write a brand new OPA policy example using the 'fairwinds' package?

The "category" field should be one of "Security", "Efficiency", or "Reliability".

The "severity" field should be a number between 0.0 and 1.0, with 1.0 indicating that the issue is very dangerous, and 0.0 indicating the issue is not bad at all.

You can use the function insightsinfo() to get more details about the context:
* insightsinfo("cluster") will return the name of the cluster being scanned
* insightsinfo("repository") will return the name of the infrastructure-as-code repository being scanned
* insightsinfo("context") will return one of "CI/CD", "AdmissionController", or "Agent"
* insightsinfo("admissionRequest") will return a Kubernetes admission object when insightsinfo("context") is AdmissionController

Here is an example policy that blocks any resources with the label "foo":
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

The new policy you write should: `


func NewClient(apiKey, model string, debug bool) *Client {
	return &Client{
		apiKey: apiKey,
		model:  model,
		debug:  debug,
	}
}

func (cs *Client) SendPrompt(ctx context.Context, prompt string) (string, error) {
	message := Message{Role: "user", Content: promptBoilerplate + prompt}
	request := Request{Model: cs.model, Messages: []Message{message}}
	content, err := sendRequest(ctx, cs.apiKey, request)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(content, "<>", ""), nil // removes the delimiters
}
