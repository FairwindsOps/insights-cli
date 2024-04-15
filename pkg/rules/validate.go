package rules

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/imroc/req"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml" // this lib correctly handles null slices
)

const rulesURLVerify = "%s/v0/organizations/%s/rules/verify-with-events"

type verifyActionItemTicketProvider string

// Defines values for VerifyActionItemTicketProvider.
const (
	VerifyActionItemTicketProviderAzure  verifyActionItemTicketProvider = "Azure"
	VerifyActionItemTicketProviderGitHub verifyActionItemTicketProvider = "GitHub"
	VerifyActionItemTicketProviderJira   verifyActionItemTicketProvider = "Jira"
)

type verifyWithEvents struct {
	ActionItem actionItem           `json:"actionItem" yaml:"actionItem"`
	Events     []ruleExecutionEvent `json:"events" yaml:"events"`
}

type ruleExecutionEvent struct {
	Changelog []changelog `json:"changelog" yaml:"changelog"`
	CreatedAt time.Time   `json:"createdAt" yaml:"createdAt"`
	Details   string      `json:"details" yaml:"details"`
	DryRun    bool        `json:"dryRun" yaml:"dryRun"`
	Level     string      `json:"level" yaml:"level"`
	Type      string      `json:"type" yaml:"type"`
}

type changelog struct {
	From any      `json:"from" yaml:"from"`
	Path []string `json:"path" yaml:"path"`
	To   any      `json:"to" yaml:"to"`
	Type string   `json:"type" yaml:"type"`
}

func (e ruleExecutionEvent) String() string {
	const colorGreen = "\033[0;32m"
	const colorYellow = "\033[0;33m"
	const colorRed = "\033[0;31m"
	const colorOrange = "\033[0;31m\033[0;33m"
	const colorNone = "\033[0m"

	var s string
	switch e.Level {
	case "error":
		s = fmt.Sprintf("%s[error]%s ", colorRed, colorNone)
	case "warn":
		s += fmt.Sprintf("%s[warn]%s ", colorYellow, colorNone)
	case "info":
		s += fmt.Sprintf("%s[info]%s ", colorGreen, colorNone)
	}

	if e.DryRun {
		s += fmt.Sprintf("%s[dry-run]%s ", colorOrange, colorNone)
	}

	switch e.Type {
	case "edit_action_item":
		var ss []string
		for _, c := range e.Changelog {
			from := genericToString(c.From)
			to := genericToString(c.To)

			ss = append(ss, fmt.Sprintf("%q was %q from %q to %q", c.Path[0], c.Type, from, to))
		}
		s += fmt.Sprintf("%q - %s - [%s]", e.Type, e.Details, strings.Join(ss, ", "))
	default:
		s += fmt.Sprintf("%q - %s", e.Type, e.Details)
	}
	return s
}

func genericToString(i any) string {
	switch v := i.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case float64:
		return fmt.Sprintf("%.2f", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

type actionItem struct {
	TicketCreatedAt   *time.Time                      `json:"TicketCreatedAt,omitempty" yaml:"TicketCreatedAt,omitempty"`
	TicketLink        *string                         `json:"TicketLink,omitempty" yaml:"TicketLink,omitempty"`
	TicketProvider    *verifyActionItemTicketProvider `json:"TicketProvider,omitempty" yaml:"TicketProvider,omitempty"`
	AssigneeEmail     *string                         `json:"assigneeEmail,omitempty" yaml:"assigneeEmail,omitempty"`
	Category          *string                         `json:"category,omitempty" yaml:"category,omitempty"`
	Cluster           *string                         `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Repository        *string                         `json:"repository,omitempty" yaml:"repository,omitempty"`
	DeletedAt         *time.Time                      `json:"deletedAt,omitempty" yaml:"deletedAt,omitempty"`
	Description       *string                         `json:"description,omitempty" yaml:"description,omitempty"`
	EventType         *string                         `json:"eventType,omitempty" yaml:"eventType,omitempty"`
	FirstSeen         *time.Time                      `json:"firstSeen,omitempty" yaml:"firstSeen,omitempty"`
	LastReportedAt    *time.Time                      `json:"lastReportedAt,omitempty" yaml:"lastReportedAt,omitempty"`
	Notes             *string                         `json:"notes,omitempty" yaml:"notes,omitempty"`
	Fixed             *bool                           `json:"fixed,omitempty" yaml:"fixed,omitempty"`
	IsCustom          *bool                           `json:"isCustom,omitempty" yaml:"isCustom,omitempty"`
	Remediation       *string                         `json:"remediation,omitempty" yaml:"remediation,omitempty"`
	ReportType        *string                         `json:"reportType,omitempty" yaml:"reportType,omitempty"`
	Resolution        *string                         `json:"resolution,omitempty" yaml:"resolution,omitempty"`
	ResourceContainer *string                         `json:"resourceContainer,omitempty" yaml:"resourceContainer,omitempty"`
	ResourceKind      *string                         `json:"resourceKind,omitempty" yaml:"resourceKind,omitempty"`
	ResourceLabels    map[string]string               `json:"resourceLabels,omitempty" yaml:"resourceLabels,omitempty"`
	ResourceName      *string                         `json:"resourceName,omitempty" yaml:"resourceName,omitempty"`
	ResourceNamespace *string                         `json:"resourceNamespace,omitempty" yaml:"resourceNamespace,omitempty"`
	Severity          *float32                        `json:"severity,omitempty" yaml:"severity,omitempty"`
	Tags              []string                        `json:"tags,omitempty" yaml:"tags,omitempty"`
	Title             *string                         `json:"title,omitempty" yaml:"title,omitempty"`
}

// ruleExecutionContext defines model for ruleExecutionContext.
type ruleExecutionContext string

// Defines values for RuleExecutionContext.
const (
	RuleExecutionContextAdmissionController ruleExecutionContext = "AdmissionController"
	RuleExecutionContextAgent               ruleExecutionContext = "Agent"
	RuleExecutionContextCICD                ruleExecutionContext = "CI/CD"
)

type verifyRule struct {
	ActionItem actionItem           `json:"actionItem" yaml:"actionItem"`
	Context    ruleExecutionContext `json:"context" yaml:"context"`
	ReportType string               `json:"reportType" yaml:"reportType"`
	Script     string               `json:"script" yaml:"script"`
}

// runVerifyRule verifies rule against one action item
func runVerifyRule(org, token, hostName string, rule verifyRule, dryRun bool) (*verifyWithEvents, error) {
	url := fmt.Sprintf(rulesURLVerify, hostName, org)
	if dryRun {
		url += "?dryRun=true"
	}

	resp, err := req.Post(url, getRuleVerifyHeaders(token), req.BodyJSON(&rule))
	if err != nil {
		return nil, fmt.Errorf("error verifying rule in Insights: %v", err)

	}
	if resp.Response().StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response code: %v - %s", resp.Response().StatusCode, string(resp.Bytes()))
	}
	var verify verifyWithEvents
	err = resp.ToJSON(&verify)
	if err != nil {
		return nil, fmt.Errorf("unable to convert response to json: %v", err)

	}
	return &verify, nil
}

func ValidateRule(org, host, insightsToken, automationRuleFilePath, actionItemFilePath, expectedActionItemFilePath, reportType, insightsContext string, dryRun bool) error {
	aiInput, err := os.Open(actionItemFilePath)
	if err != nil {
		return fmt.Errorf("error when trying to open action item file %s: %v", actionItemFilePath, err)
	}

	aiBytes, err := io.ReadAll(aiInput)
	if err != nil {
		return fmt.Errorf("could not read action item file %s: %v", actionItemFilePath, err)
	}

	var ai actionItem
	err = yaml.Unmarshal(aiBytes, &ai)
	if err != nil {
		return fmt.Errorf("could not parse action item file %s: %v", actionItemFilePath, err)
	}

	err = validateInputActionItem(insightsContext, ai)
	if err != nil {
		return fmt.Errorf("invalid input action item: %v", err)
	}

	ruleInput, err := os.Open(automationRuleFilePath)
	if err != nil {
		return fmt.Errorf("error when trying to open file %s: %v", automationRuleFilePath, err)
	}

	ruleBytes, err := io.ReadAll(ruleInput)
	if err != nil {
		return fmt.Errorf("could not read rule file %s: %v", automationRuleFilePath, err)
	}

	verifyRule := verifyRule{
		ActionItem: ai,
		Context:    ruleExecutionContext(insightsContext),
		ReportType: reportType,
		Script:     string(ruleBytes),
	}
	r, err := runVerifyRule(org, insightsToken, host, verifyRule, dryRun)
	if err != nil {
		return fmt.Errorf("unable to verify rule: %v", err)
	}
	responseActionItem := r.ActionItem

	if len(r.Events) > 0 {
		fmt.Printf("\n-- Logs --\n\n")
		for _, e := range r.Events {
			fmt.Println(e)
		}
	}

	if expectedActionItemFilePath == "" {
		fmt.Printf("\n-- Returned Action Item --\n\n")
		b, err := yaml.Marshal(responseActionItem)
		if err != nil {
			return fmt.Errorf("could not marshal verify result: %v", err)
		}
		fmt.Println(string(b))
		return nil
	}

	expectedActionItemFile, err := os.Open(expectedActionItemFilePath)
	if err != nil {
		return fmt.Errorf("error when trying to open expected file %s: %v", expectedActionItemFilePath, err)
	}

	expectedActionItemBytes, err := io.ReadAll(expectedActionItemFile)
	if err != nil {
		return fmt.Errorf("failed to read output file: %v", err)
	}

	var expectedActionItem actionItem
	err = yaml.Unmarshal(expectedActionItemBytes, &expectedActionItem)
	if err != nil {
		return fmt.Errorf("could not marshal expected response: %v", err)
	}

	fmt.Printf("\n-- Diff Result --\n\n")

	opts := buildCmpOptions(expectedActionItem)

	diff := cmp.Diff(expectedActionItem, responseActionItem, opts...)
	if len(diff) == 0 {
		logrus.Infoln("Success - actual response matches expected response")
		fmt.Println()
	} else {
		logrus.Errorln("Test failed:")
		fmt.Println(diff)
	}
	return nil
}

func validateInputActionItem(context string, ai actionItem) error {
	if ai.Title == nil || len(*ai.Title) == 0 {
		return errors.New("title is required")
	}

	if ai.EventType == nil || len(*ai.EventType) == 0 {
		return errors.New("eventType is required")
	}

	if isClusterRequired(context) {
		if ai.Cluster == nil || len(*ai.Cluster) == 0 {
			return fmt.Errorf("cluster is required for context %s", context)
		}
	}

	if isRepositoryRequired(context) {
		if ai.Repository == nil || len(*ai.Repository) == 0 {
			return fmt.Errorf("repository is required for context %s", context)
		}
	}

	return nil
}

func isRepositoryRequired(context string) bool {
	return context == string(RuleExecutionContextCICD)
}

func isClusterRequired(context string) bool {
	return context == string(RuleExecutionContextAgent) || context == string(RuleExecutionContextAdmissionController)
}

// buildCmpOptions builds the cmp options to ignore fields that are nil in the expected object
func buildCmpOptions[T any](expected T) []cmp.Option {
	var opts []cmp.Option
	typ := reflect.TypeOf(expected)
	for i := 0; i < typ.NumField(); i++ {
		fieldName := typ.Field(i).Name
		if reflect.ValueOf(expected).FieldByName(fieldName).IsNil() {
			opts = append(opts, cmpopts.IgnoreFields(expected, fieldName))
		}
	}
	return opts
}

func getRuleVerifyHeaders(token string) req.Header {
	return req.Header{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Accept":        "application/json",
		"Content-Type":  "application/yaml",
	}
}
