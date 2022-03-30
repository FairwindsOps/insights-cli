package opavalidation

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/open-policy-agent/opa/rego"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// arrayFromRegoOutput converts rego output into an array of interface{}, to
// make rego output iteratable.
func arrayFromRegoOutput(results rego.ResultSet) []interface{} {
	returnSet := make([]interface{}, 0)

	for _, result := range results {
		for _, pack := range result.Bindings["results"].(map[string]interface{}) {
			for _, outputArray := range pack.(map[string]interface{}) {
				returnSet = append(returnSet, outputArray.([]interface{})...)
			}
		}
	}
	return returnSet
}

// actionItemsFromRegoResults converts rego execution output into a slice of
// actionItem.
func actionItemsFromRegoResult(results rego.ResultSet) (actionItems, error) {
	actionItems := make(actionItems, 0)
	resultsAsArray := arrayFromRegoOutput(results)
	var allErrs *multierror.Error = new(multierror.Error)
	for n, result := range resultsAsArray {
		var AI actionItem
		resultAsMap, ok := result.(map[string]interface{})
		if ok {
			var err error
			AI, err = actionItemFromMap(resultAsMap)
			if err != nil {
				allErrs = multierror.Append(allErrs, fmt.Errorf("action item %d: %v", n+1, err))
				continue
			}
		} else {
			allErrs = multierror.Append(allErrs, fmt.Errorf("action item %d: could not decipher output format of %+v %T", n+1, result, result))
			continue
		}
		actionItems = append(actionItems, AI)
	}
	return actionItems, allErrs.ErrorOrNil()
}

// actionItemFromMap converts a map[string]interface{} to a type actionItem.
// ANy missing actionItem fields are returned in an error.
// This is used while converting rego results to actionItems.
func actionItemFromMap(m map[string]interface{}) (actionItem, error) {
	var AI actionItem
	missingFields := make(map[string]error) // Store errors from get*Field functions
	var err error
	AI.Description, err = getStringField(m, "description")
	if err != nil {
		missingFields["description"] = err
	}
	AI.Title, err = getStringField(m, "title")
	if err != nil {
		missingFields["title"] = err
	}
	AI.Category, err = getStringField(m, "category")
	if err != nil {
		missingFields["category"] = err
	}
	AI.Remediation, err = getStringField(m, "remediation")
	if err != nil {
		missingFields["remediation"] = err
	}
	AI.Severity, err = getFloatField(m, "severity")
	if err != nil {
		missingFields["severity"] = err
	}
	if len(missingFields) > 0 {
		return AI, errors.New(humanizeMapOutput(missingFields, "missing field"))
	}
	return AI, nil
}

// humanizeMapOutput returns a humanized string listing a maps keys with its
// error values
// in parenthesis. The supplied keyNoun will be pluralized if there are more
// than one key in the map.
func humanizeMapOutput(m map[string]error, keyNoun string) string {
	var message strings.Builder
	fmt.Fprintf(&message, "%d %s", len(m), keyNoun)
	if len(m) > 1 {
		message.WriteString("s")
	}
	message.WriteString(": ")
	var n int = 1 // counter of keys processed
	for k, v := range m {
		if n == len(m) && len(m) == 2 {
			message.WriteString(" and ")
		}
		if n == len(m) && len(m) > 2 {
			message.WriteString("and ") // the comma logic will provide a l eading space
		}
		fmt.Fprintf(&message, "%s (%v)", k, v)
		if n < len(m) && len(m) > 2 {
			message.WriteString(", ")
		}
		n++
	}
	return message.String()
}

// objectBytesToMap converts a slice of bytes to a map[string]interface{},
// suitable for passing into runRegoForObject().
func objectBytesToMap(objectAsBytes []byte) (map[string]interface{}, error) {
	objectAsMap := make(map[string]interface{})
	err := yaml.Unmarshal(objectAsBytes, &objectAsMap)
	if err != nil {
		return nil, fmt.Errorf("cannot process input object: %v", err)
	}
	return objectAsMap, nil
}

// updateObjectWithNamespaceOverride sets the metadata.namespace field of a
// Kubernetes object.
func updateObjectWithNamespaceOverride(obj map[string]interface{}, NS string) error {
	if NS == "" {
		return nil
	}
	err := unstructured.SetNestedField(obj, NS, "metadata", "namespace")
	if err != nil {
		return err
	}
	return nil
}

// getMapField attempts to get a sub-map at the provided key of a map[string]interface{}.
func getMapField(m map[string]interface{}, key string) (map[string]interface{}, error) {
	if m[key] == nil {
		return nil, fmt.Errorf("key %q not found", key)
	}
	subMap, ok := m[key].(map[string]interface{})
	if !ok {
		return nil, errors.New(key + " was not a map")
	}
	return subMap, nil
}

// getStringField attempts to get a string at the provided key of a
// map[string]interface{}.
func getStringField(m map[string]interface{}, key string) (string, error) {
	if m[key] == nil {
		return "", fmt.Errorf("key %q not found", key)
	}
	str, ok := m[key].(string)
	if !ok {
		return "", errors.New(key + " was not a string")
	}
	return str, nil
}

// getFloatField attempts to get a float at the provided key of a
// map[string]interface{}.
func getFloatField(m map[string]interface{}, key string) (float64, error) {
	if m[key] == nil {
		return 0.0, fmt.Errorf("key %q not found", key)
	}
	n, ok := m[key].(json.Number)
	if !ok {
		return 0.0, errors.New(key + " was not a float")
	}
	f, err := n.Float64()
	if err != nil {
		return 0.0, err
	}
	return f, nil
}
