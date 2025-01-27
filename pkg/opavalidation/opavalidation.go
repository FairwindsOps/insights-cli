// Package opavalidation validates OPA policies for Insights.
package opavalidation

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	fwrego "github.com/fairwindsops/insights-plugins/plugins/opa/pkg/rego"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/topdown"
	"github.com/open-policy-agent/opa/v1/types"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

const (
	DefaultKubeObjectNamespace = "notset" // The namespace to use if one is unspecified
)

// Run is a ValidateRego() wrapper that validates and prints resulting actionItems. This is
// meant to be called from a cobra.Command{}.
func Run(regoFileName, objectFileName string, expectAIOptions ExpectActionItemOptions, insightsInfo fwrego.InsightsInfo, objectNamespaceOverride, libsDir string) (actionItems, error) {
	b, err := os.ReadFile(regoFileName)
	if err != nil {
		return nil, fmt.Errorf("error reading OPA policy %s: %v", regoFileName, err)
	}
	regoContent := string(b)
	b, err = os.ReadFile(objectFileName)
	if err != nil {
		return nil, fmt.Errorf("error reading Kubernetes manifest %s: %v", objectFileName, err)
	}
	libs := map[string]string{}
	if libsDir != "" {
		files, err := FindFilesWithExtension(libsDir, ".rego")
		if err != nil {
			return nil, fmt.Errorf("unable to list .rego files in %s: %v", libsDir, err)
		}
		for _, lib := range files {
			libContent, err := os.ReadFile(lib)
			if err != nil {
				return nil, fmt.Errorf("error reading OPA library %s: %v", lib, err)
			}
			if !IsOPACustomLibrary(string(libContent)) {
				logrus.Warnf("Skipping non-OPA library %s", lib)
				continue
			}
			libName := strings.TrimSuffix(filepath.Base(lib), filepath.Ext(lib))
			libs[libName] = string(libContent)
		}
	}

	baseRegoFileName := filepath.Base(regoFileName)
	eventType := strings.TrimSuffix(baseRegoFileName, filepath.Ext(baseRegoFileName))
	actionItems, err := ValidateRego(context.TODO(), regoContent, b, insightsInfo, eventType, objectNamespaceOverride, libs)
	if err != nil {
		return actionItems, err
	}
	actionItemsAsString, err := actionItems.StringWithValidation()
	// If actionItems have errors, output the actionItems first to display more
	// specific inline errors.
	fmt.Println(actionItemsAsString)
	if err != nil {
		return actionItems, err
	}
	expectAI := expectAIOptions.ForFileName(objectFileName)
	if expectAI && len(actionItems) != 1 {
		return actionItems, fmt.Errorf("%d action items were returned, but 1 is expected", len(actionItems))
	}
	if !expectAI && len(actionItems) > 0 {
		return actionItems, fmt.Errorf("%d action items were returned but none are expected", len(actionItems))
	}
	return actionItems, nil
}

// RunBatch is a Run() wrapper that processes multiple OPA policies. It does
// not return the actionItems from each call to Run(), as there would not be correlation of
// actionItems to their OPA policy.
// This is meant to be called from a cobra.Command{}.
// Each OPA policy is validated with a Kubernetes manifest file named of the
// form {base rego filename} and the extensions .yaml, .success.yaml, and
// .failure.yaml (the last two of which are configurable).
func RunBatch(batchDir string, expectAIOptions ExpectActionItemOptions, insightsInfo fwrego.InsightsInfo, objectNamespaceOverride, libsDir string) (successfulPolicies, failedPolicies []string, err error) {
	regoFiles, err := FindFilesWithExtension(batchDir, ".rego")
	if err != nil {
		return successfulPolicies, failedPolicies, fmt.Errorf("unable to list .rego files: %v", err)
	}
	for _, regoFileName := range regoFiles {
		objectFileNames, ok, err := expectAIOptions.getObjectFileNamesForPolicy(regoFileName)
		if err != nil {
			return nil, nil, fmt.Errorf("error finding object files for policy %s: %w", regoFileName, err)
		}
		if !ok {
			logrus.Errorf("No Kubernetes manifest files found to use as input for validation of OPA policy %s", regoFileName)
			failedPolicies = append(failedPolicies, regoFileName)
			continue
		}
		for _, objectFileName := range objectFileNames {
			logrus.Infof("Validating OPA policy %s with input %s (expectActionItem=%v)", regoFileName, objectFileName, expectAIOptions.ForFileName(objectFileName))
			_, err := Run(regoFileName, objectFileName, expectAIOptions, insightsInfo, objectNamespaceOverride, libsDir)
			if err != nil {
				logrus.Errorf("Failed validation of OPA policy %s using input %s: %v\n", regoFileName, objectFileName, err)
				if !lo.Contains(failedPolicies, regoFileName) {
					failedPolicies = append(failedPolicies, regoFileName)
				}
			}
		}
		if !lo.Contains(failedPolicies, regoFileName) {
			successfulPolicies = append(successfulPolicies, regoFileName)
			logrus.Infof("Success validating OPA policy %s\n", regoFileName)
		}
	}
	if len(failedPolicies) > 0 {
		return successfulPolicies, failedPolicies, fmt.Errorf("%d failed  and %d succeeded", len(failedPolicies), len(successfulPolicies))
	}
	return successfulPolicies, failedPolicies, nil
}

// ValidateRego validates rego by executing rego with an input object.
// Validation includes signatures for Insights-provided rego functions.
func ValidateRego(ctx context.Context, regoAsString string, objectAsBytes []byte, insightsInfo fwrego.InsightsInfo, eventType string, objectNamespaceOverride string, libs map[string]string) (actionItems, error) {
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
	regoResult, err := runRegoForObject(ctx, regoAsString, objectAsMap, insightsInfo, libs)
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
func runRegoForObject(ctx context.Context, regoAsString string, object map[string]interface{}, insightsInfo fwrego.InsightsInfo, libs map[string]string) (rego.ResultSet, error) {
	opts := []func(r *rego.Rego){rego.EnablePrintStatements(true), rego.PrintHook(topdown.NewPrintHook(os.Stdout)),
		rego.Query("results = data"),
		rego.Module("fairwinds", regoAsString),
		rego.Function2(
			&rego.Function{
				Name: "kubernetes",
				Decl: types.NewFunction(types.Args(types.S, types.S), types.A),
			},
			func(_ rego.BuiltinContext, _ *ast.Term, _ *ast.Term) (*ast.Term, error) {
				logrus.Warnln("NOTE: The rego kubernetes function currently does not return data when validating OPA policies.")
				returnData, err := ast.InterfaceToValue([]string{"the rego kubernetes function currently does not return data when validating OPA policies"})
				if err != nil {
					return nil, err
				}
				return ast.NewTerm(returnData), nil
			},
		),
		rego.Function1(
			&rego.Function{
				Name: "insightsinfo",
				Decl: types.NewFunction(types.Args(types.S), types.A),
			},
			fwrego.GetInsightsInfoFunction(&insightsInfo),
		),
	}
	var libNames []string
	for libName := range libs {
		libNames = append(libNames, libName)
	}
	sort.Strings(libNames)
	for _, libName := range libNames {
		opts = append(opts, rego.Module(libName, libs[libName]))
	}
	query, err := rego.New(opts...).PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("error preparing rego for evaluation: %v", err)
	}
	preparedInput := rego.EvalInput(object)
	rs, err := query.Eval(ctx, preparedInput)
	if err != nil {
		return nil, fmt.Errorf("error evaluating rego: %v", err)
	}
	return rs, nil
}

var isCheckRE = regexp.MustCompile(`^package\s+fairwinds\s*(#.*)?$`)

func IsOPACustomLibrary(rego string) bool {
	for _, line := range strings.Split(strings.TrimSuffix(rego, "\n"), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "package") {
			isCheck := isCheckRE.MatchString(strings.TrimSpace(line))
			return !isCheck
		}
	}
	return false
}
