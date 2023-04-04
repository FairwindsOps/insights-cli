# Set environment for E2E tests.
# See also, our Vault path repo/insights-cli/env, which will populate this environment.

echo "Setting environment variables required by E2E tests. . ."
echo "export CI_INSIGHTS_API_URL='https://stable-main.k8s.insights.fairwinds.com'" >>${BASH_ENV}
echo "export CI_INSIGHTS_ORGANIZATION='cli-test'" >>${BASH_ENV}
