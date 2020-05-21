package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/spf13/cobra"
)

type awsCURLFlags struct {
	headers []string
	method  string
	data    string

	awsAccessKey    string
	awsSecretKey    string
	awsSessionToken string
	awsProfile      string
	awsService      string
	awsRegion       string
}

var flags awsCURLFlags

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "awscurl [URL]",
	Short: "cURL with AWS request signing",
	Long: `A simple CLI utility with cURL-like syntax allowing to send HTTP
requests to AWS resources. awscurl automatically adds an authentication information
to the HTTP request. It uses Siganture Version 4. More details:

https://docs.aws.amazon.com/general/latest/gr/signature-version-4.html
`,
	Args:    cobra.ExactArgs(1),
	RunE:    runCurl,
	Version: "0.2.0",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flags.method, "request", "X", "GET", "Custom request method to use")
	rootCmd.PersistentFlags().StringVarP(&flags.data, "data", "d", "", "Data payload to send within a POST request")
	rootCmd.PersistentFlags().StringArrayVarP(&flags.headers, "header", "H", []string{},
		`Extra HTTP header to include in the request. Example: "Content-Type: application/json" Could be used multiple times`)
	rootCmd.PersistentFlags().StringVar(&flags.awsAccessKey, "access-key", "", "AWS Access Key ID to use for authentication")
	rootCmd.PersistentFlags().StringVar(&flags.awsSecretKey, "secret-key", "", "AWS Secret Access Key to use for authentication")
	rootCmd.PersistentFlags().StringVar(&flags.awsSessionToken, "session-token", "", "AWS Session Key to use for authentication")
	rootCmd.PersistentFlags().StringVar(&flags.awsProfile, "profile", "", "AWS awsProfile to use for authentication")
	rootCmd.PersistentFlags().StringVar(&flags.awsService, "service", "execute-api", "The name of AWS Service, used for signing the request")
	rootCmd.PersistentFlags().StringVar(&flags.awsRegion, "region", "", "AWS region to use for the request")

	rootCmd.Flags().SortFlags = false
}

func runCurl(cmd *cobra.Command, args []string) error {
	// Suppress the usage info in case of errors happen below
	// We do it here, after the init(), so the usage info is still printed for invalid args and flags.
	cmd.SilenceUsage = true

	if len(args) != 1 {
		return fmt.Errorf("Error: Only one URL is expected, %d given", len(args))
	}

	cfg, err := getAWSConfig(flags)
	if err != nil {
		return err
	}

	// Build the HTTP request
	url := args[0]
	body := strings.NewReader(flags.data)
	req, err := http.NewRequest(flags.method, url, body)
	if err != nil {
		return err
	}

	for _, h := range flags.headers {
		hParts := strings.Split(h, ":")
		if len(hParts) != 2 {
			return fmt.Errorf(`Error: Invalid header: %s. It should be in the format "Name: Value"`, h)
		}
		hKey := strings.TrimSpace(hParts[0])
		hVal := strings.TrimSpace(hParts[1])
		req.Header.Add(hKey, hVal)
	}

	// Sign the HTTP request. Special headers will be added to the given *http.Request
	reqBody := readAndReplaceBody(req)
	reqBodySHA256 := hashSHA256(reqBody)
	signer := v4.NewSigner(cfg.Credentials)
	err = signer.SignHTTP(req.Context(), req, reqBodySHA256, flags.awsService, cfg.Region, time.Now())
	if err != nil {
		return err
	}

	// Semd the request and print the response
	client := new(http.Client)
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	return nil
}

// getAWSConfig builgs the AWS Config based on the provided AWS-related flags
func getAWSConfig(f awsCURLFlags) (aws.Config, error) {
	var cfg aws.Config
	var cfgSources external.Configs

	if f.awsProfile != "" {
		awsProfileLoader := external.WithSharedConfigProfile(f.awsProfile)
		cfgSources = append(cfgSources, awsProfileLoader)
	}
	if f.awsAccessKey != "" && f.awsSecretKey != "" {
		staticCredsLoader := external.WithCredentialsProvider{
			CredentialsProvider: aws.StaticCredentialsProvider{
				Value: aws.Credentials{
					AccessKeyID: f.awsAccessKey, SecretAccessKey: f.awsSecretKey, SessionToken: f.awsSessionToken,
				},
			},
		}
		cfgSources = append(cfgSources, staticCredsLoader)
	}

	cfg, err := external.LoadDefaultAWSConfig(cfgSources...)
	if err != nil {
		return cfg, fmt.Errorf("Unable to load AWS config: %s", err)
	}

	if f.awsRegion != "" {
		cfg.Region = f.awsRegion
	}

	return cfg, nil
}

func readAndReplaceBody(request *http.Request) []byte {
	if request.Body == nil {
		return []byte{}
	}
	payload, _ := ioutil.ReadAll(request.Body)
	request.Body = ioutil.NopCloser(bytes.NewReader(payload))
	return payload
}

func hashSHA256(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return fmt.Sprintf("%x", h.Sum(nil))
}
