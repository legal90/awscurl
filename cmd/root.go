package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	headers         []string
	method          string
	data            string
	awsService      string
	awsRegion       string
	awsProfile      string
	awsAccessKey    string
	awsSecretKey    string
	awsSessionToken string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "awscurl [URL]",
	Short: "cURL with AWS request signing",
	Long: `A simple CLI utility with cURL-like syntax allowing to send HTTP
requests to AWS resources. awscurl automatically adds an authentication information
to the HTTP request. It uses Siganture Version 4. More details:

https://docs.aws.amazon.com/general/latest/gr/signature-version-4.html
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCurl(cmd, args[0])
	},
	Version: "0.2.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().SortFlags = false
	defaultHeaders := []string{
		"Accept: application/xml",
		"Content-Type: application/json",
	}

	rootCmd.PersistentFlags().StringArrayVarP(&headers, "header", "H", defaultHeaders,
		"Extra HTTP header to include in the request. Could be used multiple times")

	rootCmd.PersistentFlags().StringVarP(&method, "request", "X", "GET", "Custom request method to use")
	rootCmd.PersistentFlags().StringVarP(&data, "data", "d", "", "Data payload to send within a POST request")

	rootCmd.PersistentFlags().StringVar(&awsService, "service", "execute-api", "The name of AWS Service, used for signing the request")

	rootCmd.PersistentFlags().StringVar(&awsRegion, "region", "", "AWS region to use for authentication")
	rootCmd.PersistentFlags().StringVar(&awsProfile, "profile", "", "AWS profile to use for authentication")
	rootCmd.PersistentFlags().StringVar(&awsAccessKey, "access-key", "", "AWS Access Key ID to use for authentication")
	rootCmd.PersistentFlags().StringVar(&awsSecretKey, "secret-key", "", "AWS Secret Access Key to use for authentication")
	rootCmd.PersistentFlags().StringVar(&awsSessionToken, "session-token", "", "AWS Session Key to use for authentication")
}

func runCurl(cmd *cobra.Command, url string) error {
	fmt.Println("Running awscurl with:", url)
	return nil
}
