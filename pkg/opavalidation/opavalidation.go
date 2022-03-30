// Package opavalidation validates OPA policies for Insights.
package opavalidation

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	fwrego "github.com/fairwindsops/insights-plugins/plugins/opa/pkg/rego"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown"
	"github.com/open-policy-agent/opa/types"
)

// Run is a ValidateRego() wrapper that prints resulting actionItems. This is
// meant to be called from a cobra.Command{}.
func Run(regoFileName, objectFileName string, insightsInfo fwrego.InsightsInfo, objectNamespaceOverride string) error {
	b, err := ioutil.ReadFile(regoFileName)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", regoFileName, err)
	}
	regoContent := string(b)
	b, err = ioutil.ReadFile(objectFileName)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", objectFileName, err)
	}
	baseRegoFileName := filepath.Base(regoFileName)
	eventType := strings.TrimSuffix(baseRegoFileName, filepath.Ext(baseRegoFileName))
	actionItems, err := ValidateRego(context.TODO(), regoContent, b, insightsInfo, eventType, objectNamespaceOverride)
	if err != nil {
		return err
	}
	actionItemsAsString, err := actionItems.StringWithValidation()
	// If actionItems have errors, output the actionItems first to display more
	// specific inline errors.
	fmt.Println(actionItemsAsString)
	if err != nil {
		return err
	}
	return nil
}

// ValidateRego validates rego by executing rego with an input object.
// Validation includes signatures for Insights-provided rego functions.
func ValidateRego(ctx context.Context, regoAsString string, objectAsBytes []byte, insightsInfo fwrego.InsightsInfo, eventType string, objectNamespaceOverride string) (actionItems, error) {
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
	regoResult, err := runRegoForObject(ctx, regoAsString, objectAsMap, insightsInfo)
	if err != nil {
		return nil, err
	}
	actionItems, err := actionItemsFromRegoResult(regoResult)
	if err != nil {
		return actionItems, err
	}
	err = actionItems.setFieldsFromObject(objectAsMap)
	if err != nil {
		return actionItems, err
	}
	actionItems.setEventType(eventType)
	return actionItems, nil
}

// runRegoForObject executes rego with a Kubernetes object as input.
func runRegoForObject(ctx context.Context, regoAsString string, object map[string]interface{}, insightsInfo fwrego.InsightsInfo) (rego.ResultSet, error) {
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
			fwrego.GetInsightsInfoFunction(&insightsInfo))).PrepareForEval(ctx)
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
