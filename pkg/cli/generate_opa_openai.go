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

const (
	prePolicyNotice = "NOTICE: The OpenAI integration is available for your convenience. Please be aware that you are using your OpenAI API key and all interaction will be governed by your agreement with OpenAI."
	inPolicyNotice  = "# NOTICE: This policy was generated in part with OpenAI’s large-scale language-generation model. The generated policy should be reviewed and tested for accuracy and revised in order to obtain the desired outcome. The User is responsible for the accuracy of the policy."
)

var openAIAPIKey, openAIModel, openAIPrompt string

// insights-cli generate opa open-ai
func init() {
	generateOPAOpenAI.Flags().StringVarP(&openAIAPIKey, "openai-api-key", "k", "", "The API key for OpenAI")
	generateOPAOpenAI.Flags().StringVarP(&openAIModel, "openai-model", "m", "gpt-4", "The OpenAI model to use")
	generateOPAOpenAI.Flags().StringVarP(&openAIPrompt, "prompt", "p", "", "A text specification of the desired OPA policy")
	generateOPACmd.AddCommand(generateOPAOpenAI)
}

// generateOPAOpenAI represents the validate opa command
var generateOPAOpenAI = &cobra.Command{
	Use:   "openai [-m model] [-k key] -p prompt",
	Short: "generate an Insights OPA policy with the help of OpenAI's ChatGPT.",
	Long:  `use OpenAI's ChatGPT API to generate an OPA policy by providing an English-language prompt describing what the policy should do. You will need to provide your own OpenAI API key.`,
	Example: `
	To generate a policy that blocks anyone from using the default namespace: insights-cli generate opa openai -k $OPENAI_API_KEY -m gpt-4 -p "blocks anyone from using the default namespace"
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(prePolicyNotice)
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
		fmt.Printf("\n\n%s\n%s\n", inPolicyNotice, content)
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
