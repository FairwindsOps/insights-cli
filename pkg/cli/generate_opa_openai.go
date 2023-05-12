package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/fairwindsops/insights-cli/pkg/openai"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var openAIAPIKey, openAIModel, openAIPrompt string

// insights-cli generate opa open-api
func init() {
	generateOPAOpenAI.Flags().StringVarP(&openAIAPIKey, "openai-api-key", "k", "", "The API key for OpenAI")
	generateOPAOpenAI.Flags().StringVarP(&openAIModel, "openai-model", "m", "gpt-4", "The OpenAI model to use")
	generateOPAOpenAI.Flags().StringVarP(&openAIPrompt, "prompt", "p", "", "A text specification of the desired OPA policy")
	generateOPACmd.AddCommand(generateOPAOpenAI)
}

// generateOPAOpenAI represents the validate opa command
var generateOPAOpenAI = &cobra.Command{
	Use:   "openai [-m model] [-k key] -p prompt",
	Short: "generate an Insights OPA policy based on OpenAI.",
	Long:  `generate an Insights OPA policy based on OpenAI.`,
	Example: `
	To generate a policy that blocks anyone from using the default namespace: insights-cli generate opa openai -k $OPENAI_API_KEY -m gpt-4 -p "blocks anyone from using the default namespace"
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if !checkGenerateOPAFlags() {
			err := cmd.Help()
			if err != nil {
				logrus.Error(err)
			}
			os.Exit(1)
		}
		apiKey := os.Getenv("OPENAI_API_KEY")
		if len(openAIAPIKey) != 0 {
			apiKey = openAIAPIKey
		}
		client := openai.NewClient(apiKey, openAIModel, true)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		content, err := client.SendPrompt(ctx, openAIPrompt)
		if err != nil {
			logrus.Fatal(err)
		}
		fmt.Printf("\n%s\n", content)
	},
}

// checkGenerateOPAFlags verifies supplied flags for `generate opa` are valid.
func checkGenerateOPAFlags() bool {
	apiKeyFromEnv := os.Getenv("OPENAI_API_KEY")
	if len(openAIAPIKey) == 0 && len(apiKeyFromEnv) == 0 {
		logrus.Errorln("Please specify one of the --openai-api-key or provide it via environment variable OPENAI_API_KEY.")
		return false
	}
	if len(openAIPrompt) == 0 {
		logrus.Errorln("Please specify --prompt flag.")
		return false
	}
	return true
}
