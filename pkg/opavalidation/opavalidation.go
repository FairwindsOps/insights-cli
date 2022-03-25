// Package opavalidation validates OPA policies for Insights.
package opavalidation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/go-multierror"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown"
	"github.com/open-policy-agent/opa/types"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Type actionItem represents an Insights action item.
type actionItem struct {
	ResourceNamespace string
	ResourceKind      string
	ResourceName      string
	Title             string
	Description       string
	Remediation       string
	EventType         string
	Severity          float64
	Category          string
}

// Type actionItems represents multiple actionItem types.
type actionItems []actionItem

// StringWithValidation validates and returns a string representation of the slice of actionItems, including
// errors with each invalid actionItem. IF any ActionItems are invalid, a
// general error is returned stating how many were invalid, but
// per-action-item errors are contained in the returned string output for
// context.
func (AIs actionItems) StringWithValidation() (string, error) {
	var numInvalidActionItems int
	var sb strings.Builder

	for n, AI := range AIs {
		valid, AIErrs := AI.valid()
		if !valid {
			numInvalidActionItems++
			sb.WriteString(color.RedString("X "))
		} else {
			sb.WriteString(color.GreenString("✔ "))
		}
		sb.WriteString("Action Item")
		if len(AIs) > 1 {
			fmt.Fprintf(&sb, " #%d", n+1)
		}
		fmt.Fprintf(&sb, ":\n")

		fmt.Fprintf(&sb, `	Title: %s
	Category: %s
	Severity: %.1f
	Description: %s
	Resource Namespace: %s
	Resource Kind: %s
	Resource Name: %s
	Remediation: %s
	Event Type: %s`,
			AI.Title, AI.Category, AI.Severity, AI.Description, AI.ResourceNamespace, AI.ResourceKind, AI.ResourceName, AI.Remediation, AI.EventType)
		fmt.Fprintln(&sb)
		if !valid {
			fmt.Fprintln(&sb, strings.TrimSpace(AIErrs.Error())) // hashicorp/multierror adds too many newlines
		}
		fmt.Fprintln(&sb)
	}
	// Per-actionItem errors were already displayed above.
	if numInvalidActionItems == 1 {
		return sb.String(), errors.New("1 action item is invalid")
	}
	if numInvalidActionItems > 1 {
		return sb.String(), fmt.Errorf("%d action items are invalid", numInvalidActionItems)
	}
	return sb.String(), nil
}

// NumInvalid returns the number of actionItems that do not pass validation.
func (AIs actionItems) NumInvalid() int {
	var numInvalidActionItems int
	for _, AI := range AIs {
		valid, _ := AI.valid()
		if !valid {
			numInvalidActionItems++
		}
	}
	return numInvalidActionItems
}

// Valid returns true if an actionItem has required fields set, or false and
// an error describing which fields are missing.
func (AI actionItem) valid() (bool, error) {
	var allErrs *multierror.Error = new(multierror.Error)
	if AI.ResourceKind == "" {
		allErrs = multierror.Append(allErrs, errors.New("ResourceKind is not set. Perhaps the input Kubernetes manifest has no `kind` field?"))
	}
	if AI.ResourceName == "" {
		allErrs = multierror.Append(allErrs, errors.New("ResourceName is not set. Perhaps the input Kubernetes manifest has no `metadata.name` field?"))
	}
	if AI.Title == "" {
		allErrs = multierror.Append(allErrs, errors.New("Title is not set."))
	}
	categoryIsValid := AI.Category == "Efficiency" ||
		AI.Category == "Security" ||
		AI.Category == "Reliability"
	if !categoryIsValid {
		allErrs = multierror.Append(allErrs, fmt.Errorf("Category %q is invalid. Category must be set to one of Efficiency, Security, or Reliability, including the uppercase first letter.", AI.Category))
	}
	return (allErrs.Len() == 0), allErrs.ErrorOrNil()
}

// setFieldsFromobject sets the resourceName, resourceNamespace, and resourceKind,
// actionItem fields from a Kubernetes object.
func (AI *actionItem) setFieldsFromObject(obj map[string]interface{}) error {
	objKind, err := getStringField(obj, "kind")
	if err != nil {
		return fmt.Errorf("while getting the Kind from the Kubernetes object: %v", err)
	}
	AI.ResourceKind = objKind
	objMetadata, err := getMapField(obj, "metadata")
	if err != nil {
		return fmt.Errorf("while getting the metadata from the Kubernetes object: %v", err)
	}
	objName, err := getStringField(objMetadata, "name")
	if err != nil {
		return fmt.Errorf("while getting the metadata.name from the Kubernetes object: %v", err)
	}
	AI.ResourceName = objName
	objNamespace, err := getStringField(objMetadata, "namespace")
	if err != nil {
		return fmt.Errorf("while getting the metadata.namespace from the Kubernetes object: %v", err)
	}
	AI.ResourceNamespace = objNamespace
	return nil
}

// Run is a ValidateRego() wrapper that prints resulting actionItems. This is
// meant to be called from a cobra.Command{}.
func Run(regoFileName, objectFileName, objectNamespaceOverride string) error {
	b, err := ioutil.ReadFile(regoFileName)
	if err != nil {
		return fmt.Errorf("error reading %s: %v\n", regoFileName, err)
	}
	regoContent := string(b)
	b, err = ioutil.ReadFile(objectFileName)
	if err != nil {
		return fmt.Errorf("error reading %s: %v\n", objectFileName, err)
	}
	baseRegoFileName := filepath.Base(regoFileName)
	eventType := strings.TrimSuffix(baseRegoFileName, filepath.Ext(baseRegoFileName))
	actionItems, err := ValidateRego(context.TODO(), regoContent, b, eventType, objectNamespaceOverride)
	if err != nil {
		return fmt.Errorf("Policy failed validation: %v\n", err)
	}
	actionItemsAsString, err := actionItems.StringWithValidation()
	fmt.Println(actionItemsAsString)
	if err != nil {
		return fmt.Errorf("Policy failed validation: %v\n", err)
	}
	fmt.Println("Policy validated successfully.")
	return nil
}

// ValidateRego validates rego by executing rego with an input object.
// Validation includes signatures for Insights-provided rego functions.
func ValidateRego(ctx context.Context, regoAsString string, objectAsBytes []byte, eventType string, objectNamespaceOverride string) (actionItems, error) {
	if !strings.Contains(regoAsString, "package fairwinds") {
		return nil, errors.New("policy must be within a fairwinds package. The policy must contain the statement: package fairwinds")
	}
	objectAsMap, err := objectBytesToMap(objectAsBytes)
	if err != nil {
		return nil, err
	}
	err = updateObjectWithNamespaceOverride(objectAsMap, objectNamespaceOverride)
	if err != nil {
		return nil, fmt.Errorf("while overriding object namespace with %q: %v", objectNamespaceOverride, err)
	}
	regoResult, err := runRegoForObject(ctx, regoAsString, objectAsMap)
	if err != nil {
		return nil, err
	}
	actionItems, err := actionItemsFromRegoResult(regoResult)
	if err != nil {
		return actionItems, err
	}
	err = updateActionItemsWithObjectFields(actionItems, objectAsMap)
	if err != nil {
		return actionItems, err
	}
	updateActionItemsWithEventType(actionItems, eventType)
	return actionItems, nil
}

// runRegoForObject executes rego with a Kubernetes object as input.
func runRegoForObject(ctx context.Context, regoAsString string, object map[string]interface{}) (rego.ResultSet, error) {
	query, err := rego.New(rego.EnablePrintStatements(true), rego.PrintHook(topdown.NewPrintHook(os.Stdout)),
		rego.Query("results = data"),
		rego.Module("fairwinds", regoAsString),
		rego.Function2(
			&rego.Function{
				Name: "kubernetes",
				Decl: types.NewFunction(types.Args(types.S, types.S), types.A),
			},
			func(_ rego.BuiltinContext, _ *ast.Term, _ *ast.Term) (*ast.Term, error) {
				// Perhaps do something mroe here to communicate it isn't possible to fetch
				// cluster resources?
				return nil, nil
			},
		),
		rego.Function1(
			&rego.Function{
				Name: "insightsinfo",
				Decl: types.NewFunction(types.Args(types.S), types.A),
			},
			validateInsightsInfoFunctionArgs())).PrepareForEval(ctx)
	if err != nil {
		return nil, err
	}
	preparedInput := rego.EvalInput(object)
	rs, err := query.Eval(ctx, preparedInput)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

// validateInsightsInfoFunctionArgs validates the argument that would be
// passed to our rego InsightsInfo function without executing code provided by
// the real InsightsInfo function used elsewhere in Insights.
// This function returns a function that is called from rego.Function().
func validateInsightsInfoFunctionArgs() func(rego.BuiltinContext, *ast.Term) (*ast.Term, error) {
	return func(bc rego.BuiltinContext, inf *ast.Term) (*ast.Term, error) {
		reqInfo, err := getStringFromAST(inf)
		if err != nil {
			return nil, rego.NewHaltError(fmt.Errorf("unable to convert requested InsightsInfo to string: %w", err))
		}
		switch strings.ToLower(reqInfo) {
		case "context", "cluster":
			// Actually return an ast string instead?
			return nil, nil
		default:
			return nil, rego.NewHaltError(fmt.Errorf("cannot return unknown Insights Info %q", reqInfo))
		}
		/*
		   retInfoAsValue, err := ast.InterfaceToValue(retInfo)
		   		if err != nil {
		   			return nil, rego.NewHaltError(fmt.Errorf("unable to convert information %q to ast value: %w", retInfo, err))
		   		}
		   		return ast.NewTerm(retInfoAsValue), nil
		*/
		return nil, nil
	}
}

// getStringFromAST converts the ast.Term type used by the rego pkg, into a
// string.
func getStringFromAST(astTerm *ast.Term) (string, error) {
	astString, ok := astTerm.Value.(ast.String)
	if !ok {
		return "", errors.New("Expected a string")
	}
	return strings.Trim(astString.String(), "\""), nil
}

// arrayFromRegoOutput converts rego output into an array of interface{}, to
// make rego output iteratable.
func arrayFromRegoOutput(results rego.ResultSet) []interface{} {
	returnSet := make([]interface{}, 0)

	for _, result := range results {
		for _, pack := range result.Bindings["results"].(map[string]interface{}) {
			for _, outputArray := range pack.(map[string]interface{}) {
				for _, output := range outputArray.([]interface{}) {
					returnSet = append(returnSet, output)
				}
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
	for _, result := range resultsAsArray {
		var AI actionItem
		resultAsMap, ok := result.(map[string]interface{})
		if ok {
			var err error
			AI, err = actionItemFromMap(resultAsMap)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("could not decipher output format of %+v %T", result, result)
		}
		actionItems = append(actionItems, AI)
	}
	return actionItems, nil
}

// actionItemFromMap converts a map[string]interface{} to a type actionItem.
// This is used while converting rego results to actionItems.
func actionItemFromMap(m map[string]interface{}) (actionItem, error) {
	var AI actionItem
	var err error
	AI.Description, err = getStringField(m, "description")
	if err != nil {
		return AI, fmt.Errorf("while getting description from action item: %v", err)
	}
	AI.Title, err = getStringField(m, "title")
	if err != nil {
		return AI, fmt.Errorf("while getting title from action item: %v", err)
	}
	AI.Category, err = getStringField(m, "category")
	if err != nil {
		return AI, fmt.Errorf("while getting category from action item: %v", err)
	}
	AI.Remediation, err = getStringField(m, "remediation")
	if err != nil {
		return AI, fmt.Errorf("while getting remediation from action item: %v", err)
	}
	AI.Severity, err = getFloatField(m, "severity")
	if err != nil {
		return AI, fmt.Errorf("while getting severity from action item: %v", err)
	}
	return AI, nil
}

// objectBytesToMap converts a slice of bytes to a map[string]interface{},
// suitable for passing into runRegoForObject().
func objectBytesToMap(objectAsBytes []byte) (map[string]interface{}, error) {
	objectAsMap := make(map[string]interface{}, 0)
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

// updateActionItemsWithObjectFields adds the Kind, name, and namespace of a
// Kubernetes object to  all actionItems in the slice of actionItems.
func updateActionItemsWithObjectFields(AIs actionItems, obj map[string]interface{}) error {
	for n, _ := range AIs {
		err := AIs[n].setFieldsFromObject(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

// updateActionItemsWithEventType adds the EventType to all actionItems in the
// slice of actionItems.
func updateActionItemsWithEventType(AIs actionItems, ET string) {
	if ET == "" {
		return
	}
	for n, _ := range AIs {
		AIs[n].EventType = ET
	}
}

// getMapField attempts to get a sub-map at the provided key of a map[string]interface{}.
func getMapField(m map[string]interface{}, key string) (map[string]interface{}, error) {
	if m[key] == nil {
		return nil, fmt.Errorf("key %q not found", key)
	}
	subMap := make(map[string]interface{}, 0)
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

// getStringField attempts to get a float at the provided key of a
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
