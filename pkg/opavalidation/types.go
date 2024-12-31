package opavalidation

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
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
		allErrs = multierror.Append(allErrs, errors.New("title is not set"))
	}
	categoryIsValid := AI.Category == "Efficiency" ||
		AI.Category == "Security" ||
		AI.Category == "Reliability"
	if !categoryIsValid {
		allErrs = multierror.Append(allErrs, fmt.Errorf("category %q is invalid. Category must be set to one of Efficiency, Security, or Reliability, including the uppercase first letter", AI.Category))
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
		logrus.Debugf("Using namespace %q, as getting the metadata.namespace from the Kubernetes object returned: %v", DefaultKubeObjectNamespace, err)
		objNamespace = DefaultKubeObjectNamespace
	}
	AI.ResourceNamespace = objNamespace
	return nil
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

// setFieldsFromObject adds the Kind, name, and namespace of a
// Kubernetes object to  all actionItems in the slice of actionItems.
func (AIs actionItems) setFieldsFromObject(obj map[string]interface{}) error {
	for n := range AIs {
		err := AIs[n].setFieldsFromObject(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

// setEventType adds the EventType to all actionItems in the
// slice of actionItems.
func (AIs actionItems) setEventType(ET string) {
	if ET == "" {
		return
	}
	for n := range AIs {
		AIs[n].EventType = ET
	}
}

// ExpectActionItemOptions bundles multiple settings about whether and when
// OPA policies are expected to output an action item.
// If a Kubernetes manifest file has the SuccessFileExtension, no action item
// is expected.
// If a Kubernetes manifest file has the FailureFileExtension, an action item
// is expected.
// The default expectation of an action item is used when the Kubernetes
// manifest file has neither file extension.
type ExpectActionItemOptions struct {
	Default                                    bool // Used if none of the below filename extensions applies.
	SuccessFileExtension, FailureFileExtension string
}

// ForFileName returns true if the given Kubernetes manifest file name should
// expectan OPA policy to output an action item.
func (o ExpectActionItemOptions) ForFileName(fileName string) bool {
	LCFileName := strings.ToLower(fileName)
	LCSuccessExtension := strings.ToLower(o.SuccessFileExtension)
	LCFailureExtension := strings.ToLower(o.FailureFileExtension)
	if strings.HasSuffix(LCFileName, LCSuccessExtension) {
		logrus.Debugf("ExpectActionItem=%v for Kube manifest file %s due to its file extension %q", false, fileName, o.SuccessFileExtension)
		return false
	}
	if strings.HasSuffix(LCFileName, LCFailureExtension) {
		logrus.Debugf("ExpectActionItem=%v for Kube manifest file %s due to its file extension %q", true, fileName, o.FailureFileExtension)
		return true
	}
	logrus.Debugf("ExpectActionItem=%v for Kube manifest file %s due to the default or the command-line flag.", o.Default, fileName)
	return o.Default
}

// getObjectFileNamesForPolicy returns a list of existing file names matching
// the pattern {base rego file name}.yaml|.success.yaml|.failure.yaml (the
// latter two being configurable via the expectActionItemOptions struct).
func (o ExpectActionItemOptions) getObjectFileNamesForPolicy(batchDir, regoFileName string) (objectFileNames []string, foundAny bool, err error) {
	baseFileName := strings.TrimSuffix(regoFileName, filepath.Ext(regoFileName))

	filenameYaml := fmt.Sprintf(`^%s\.yaml$`, regexp.QuoteMeta(baseFileName))
	anythingFailureYaml := fmt.Sprintf(`^%s(\.[a-zA-Z0-9_-]+)?\.failure\.yaml$`, regexp.QuoteMeta(baseFileName))
	anythingSuccessYaml := fmt.Sprintf(`^%s(\.[a-zA-Z0-9_-]+)?\.success\.yaml$`, regexp.QuoteMeta(baseFileName))

	filenameYamlRegex := regexp.MustCompile(filenameYaml)
	anythingFailureYamlRegex := regexp.MustCompile(anythingFailureYaml)
	anythingSuccessYamlRegex := regexp.MustCompile(anythingSuccessYaml)

	files, err := ListAllFilesInDir(batchDir, true)
	if err != nil {
		return nil, false, err
	}

	for _, file := range files {
		if filenameYamlRegex.MatchString(file) {
			objectFileNames = append(objectFileNames, file)
		}
		if anythingFailureYamlRegex.MatchString(file) {
			objectFileNames = append(objectFileNames, file)
		}
		if anythingSuccessYamlRegex.MatchString(file) {
			objectFileNames = append(objectFileNames, file)
		}
	}

	if len(objectFileNames) > 0 {
		logrus.Debugf("Matched these object files for policy %s: %v", regoFileName, objectFileNames)
		foundAny = true
	}

	return
}
