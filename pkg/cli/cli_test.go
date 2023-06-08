//go:build e2e

// Use the testscript package to affectively execute the insights-cli binary
// for E2E tests.
// Ref: https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript
package cli_test

// Note for troubleshooting: adding the -testwork flag to the go test command,
// will preserve the TestScript temporary directory for manual examination.

import (
	"fmt"
	"os"
	"testing"

	"github.com/fairwindsops/insights-cli/pkg/cli"

	"github.com/rogpeppe/go-internal/testscript"
)

// A TestScript setup function that passes specific environment variables into
// the tests.
var testScriptSetup func(*testscript.Env) error = func(e *testscript.Env) error {
	e.Vars = append(e.Vars,
		fmt.Sprintf("FAIRWINDS_TOKEN=%s", os.Getenv("FAIRWINDS_TOKEN")),
		fmt.Sprintf("CI_INSIGHTS_API_URL=%s", os.Getenv("CI_INSIGHTS_API_URL")),
		fmt.Sprintf("CI_INSIGHTS_ORGANIZATION=%s", os.Getenv("CI_INSIGHTS_ORGANIZATION")))
	return nil
}

func TestMain(m *testing.M) {
	// Map running `insights-cli`, to call our main function.
	// This avoids tests running an unexpected insights-cli that may be in $PATH
	// on a local workstation.
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"insights-cli": cli.Run,
	}))
}

// Define a test that runs TestScript to process script files in
// the testdata/scripts directory.
func TestScript(t *testing.T) {
	t.Parallel()
	testscript.Run(t, testscript.Params{
		Dir:   "../../testdata/scripts",
		Setup: testScriptSetup,
	})
}
