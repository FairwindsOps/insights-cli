// Package opavalidation validates OPA policies for Insights.
package opavalidation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown"
	"github.com/open-policy-agent/opa/types"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ActionItem represents an action item from a report
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

type actionItems []actionItem

// StringWithValidation validates and prints the slice of actionItems, including
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
			sb.WriteString("X ")
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
	Event Type: %s
	`, AI.Title, AI.Category, AI.Severity, AI.Description, AI.ResourceNamespace, AI.ResourceKind, AI.ResourceName, AI.Remediation, AI.EventType)
		if !valid {
			fmt.Fprintln(&sb, AIErrs)
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

// setFieldsFromobject sets the resourceName, resourceNamespace, and resourceKind,
// actionItem fields from a Kubernetes object.
func (AI *actionItem) setFieldsFromObject(obj map[string]interface{}) error {
	objKind, err := getStringField(obj, "kind")
	if err != nil {
		return err
	}
	AI.ResourceKind = objKind
	objMetadata, err := getMapField(obj, "metadata")
	if err != nil {
		return err
	}
	objName, err := getStringField(objMetadata, "name")
	if err != nil {
		return fmt.Errorf("while parsing metadata: %v", err)
	}
	AI.ResourceName = objName
	objNamespace, err := getStringField(objMetadata, "namespace")
	if err != nil {
		return fmt.Errorf("while parsing metadata: %v", err)
	}
	AI.ResourceNamespace = objNamespace
	return nil
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

// ValidateRego validates rego by parsing inputs, modules, and query arguments
// via the upstream rego pkg. This includes validating signatures for
// Insights-provided rego functions.
func ValidateRego(ctx context.Context, regoAsString string, objectAsBytes []byte, eventType string, objectNamespaceOverride string) error {
	if !strings.Contains(regoAsString, "package fairwinds") {
		return errors.New("policy must be within a fairwinds package. The policy must contain the statement: package fairwinds")
	}
	objectAsMap, err := objectBytesToMap(objectAsBytes)
	if err != nil {
		return err
	}
	err = updateObjectWithNamespaceOverride(objectAsMap, objectNamespaceOverride)
	if err != nil {
		return fmt.Errorf("while overriding object namespace with %q: %v", objectNamespaceOverride, err)
	}
	regoResult, err := runRegoForObject(ctx, regoAsString, objectAsMap)
	if err != nil {
		return err
	}
	actionItems, err := actionItemsFromRegoResult(regoResult)
	if err != nil {
		return err
	}
	err = updateActionItemsWithObjectFields(actionItems, objectAsMap)
	if err != nil {
		return err
	}
	updateActionItemsWithEventType(actionItems, eventType)
	actionItemsAsString, err := actionItems.StringWithValidation()
	fmt.Println(actionItemsAsString)
	if err != nil {
		return err
	}
	return nil
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
// the real InsightsInfo function. This function returns a function that is
// called from rego.Function().
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

func getStringFromAST(astTerm *ast.Term) (string, error) {
	astString, ok := astTerm.Value.(ast.String)
	if !ok {
		return "", errors.New("Expected a string")
	}
	return strings.Trim(astString.String(), "\""), nil
}

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
		return AI, err
	}
	AI.Title, err = getStringField(m, "title")
	if err != nil {
		return AI, err
	}
	AI.Category, err = getStringField(m, "category")
	if err != nil {
		return AI, err
	}
	AI.Remediation, err = getStringField(m, "remediation")
	if err != nil {
		return AI, err
	}
	AI.Severity, err = getFloatField(m, "severity")
	if err != nil {
		return AI, err
	}
	return AI, nil
}

func objectBytesToMap(objectAsBytes []byte) (map[string]interface{}, error) {
	objectAsMap := make(map[string]interface{}, 0)
	err := yaml.Unmarshal(objectAsBytes, &objectAsMap)
	if err != nil {
		return nil, fmt.Errorf("cannot process input object: %v", err)
	}
	return objectAsMap, nil
}

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

func updateActionItemsWithObjectFields(AIs actionItems, obj map[string]interface{}) error {
	for n, _ := range AIs {
		err := AIs[n].setFieldsFromObject(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateActionItemsWithEventType(AIs actionItems, ET string) {
	if ET == "" {
		return
	}
	for n, _ := range AIs {
		AIs[n].EventType = ET
	}
}

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
