# Command-line Test Scripts

This directory contains test scripts that are processed by [the Go TestScript package](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript). This is used to test end-to-end behavior of the Insights CLI by essentially running the `insights-cli` command, then checking its output.

Here are a few notes about the "scripts" in this directory:

* Scripts are interpreted by TestScript, not by a shell like bash.
* Scripts are specified in [text archive formatted .txtar files](https://pkg.go.dev/golang.org/x/tools/txtar) - a text file that contains the test script, as well as other inline files that the test requires.
* Each script runs in its own temporary directory.
* The `stdout` and `stderr` script commands use a regular expression to match text.
* To see more about other available commands that can be used in scripts, see the "The predefined commands are:" section of [the TestScript docs](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript).

* See the [pkg/cli/test_cli.go file](./../../pkg/cli/cli_test.go) whichs configures and executes TestScript to process these script files.
