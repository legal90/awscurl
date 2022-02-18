package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	urls "net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
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
	insecure        bool
	proxy           string
}

var (
	// Version and git commit SHA to include to the `--version` output
	// These variables are supposed to be overriden on the build time using ldflags
	version = "dev"
	commit  = "none"

	flags awsCURLFlags
)

// rootCmd represents the base awscurl command when called without any subcommands (which we don't have here)
var rootCmd = &cobra.Command{
	Use:   "awscurl [URL]",
	Short: "cURL with AWS request signing",
	Long: `A simple CLI utility with cURL-like syntax allowing to send HTTP requests to AWS resources.
It automatically adds Signature Version 4 to the request. More details:
https://docs.aws.amazon.com/general/latest/gr/signature-version-4.html
`,
	Args:    cobra.ExactArgs(1),
	RunE:    runCurl,
	Version: fmt.Sprintf("%s, build %s", version, commit),
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flags.method, "request", "X", "GET", "Custom request method to use")
	rootCmd.PersistentFlags().StringVarP(&flags.data, "data", "d", "", `Data payload to send within a request. Could be also read from a file if prefixed with @, example: -d "@/path/to/file.json"`)
	rootCmd.PersistentFlags().StringArrayVarP(&flags.headers, "header", "H", []string{},
		`Extra HTTP header to include in the request. Example: -h "Content-Type: application/json". Could be used multiple times`)
	rootCmd.PersistentFlags().StringVar(&flags.awsAccessKey, "access-key", "", "AWS Access Key ID to use for authentication")
	rootCmd.PersistentFlags().StringVar(&flags.awsSecretKey, "secret-key", "", "AWS Secret Access Key to use for authentication")
	rootCmd.PersistentFlags().StringVar(&flags.awsSessionToken, "session-token", "", "AWS Session Key to use for authentication")
	rootCmd.PersistentFlags().StringVar(&flags.awsProfile, "profile", "", "AWS awsProfile to use for authentication")
	rootCmd.PersistentFlags().StringVar(&flags.awsService, "service", "execute-api", "The name of AWS Service, used for signing the request")
	rootCmd.PersistentFlags().StringVar(&flags.awsRegion, "region", "", "AWS region to use for the request")
	rootCmd.PersistentFlags().BoolVarP(&flags.insecure, "insecure", "k", false, "Allow insecure server connections when using SSL")
	rootCmd.PersistentFlags().StringVar(&flags.proxy, "proxy", "", "Proxy to use for the request")

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

	var body io.Reader

	if strings.HasPrefix(flags.data, "@") {
		// Read data from file
		fPath := flags.data[1:]
		body, err = os.Open(fPath)
		if err != nil {
			return err
		}
	} else {
		body = strings.NewReader(flags.data)
	}

	// Build the HTTP request
	url := args[0]
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
	signer := v4.NewSigner()

	creds, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		return err
	}

	err = signer.SignHTTP(req.Context(), creds, req, reqBodySHA256, flags.awsService, cfg.Region, time.Now())
	if err != nil {
		return err
	}

	// Set TLS Client configuration
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: flags.insecure},
	}

	// Add proxy if needed
	if flags.proxy != "" {
		//creating the proxyURL
		proxyURL, err := urls.Parse(flags.proxy)
		if err != nil {
			return err
		}
		//adding the proxy settings to the Transport object
		tr.Proxy = http.ProxyURL(proxyURL)
	}

	// Send the request and print the response
	client := http.Client{Transport: tr}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var content []byte
	content, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(content))

	return nil
}

// getAWSConfig builgs the AWS Config based on the provided AWS-related flags
func getAWSConfig(f awsCURLFlags) (aws.Config, error) {
	var cfg aws.Config
	var cfgSources []func(*config.LoadOptions) error

	if f.awsProfile != "" {
		awsProfileLoader := config.WithSharedConfigProfile(f.awsProfile)
		cfgSources = append(cfgSources, awsProfileLoader)
	}
	if f.awsAccessKey != "" && f.awsSecretKey != "" {
		staticCredsLoader := config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(f.awsAccessKey, f.awsSecretKey, f.awsSessionToken))
		cfgSources = append(cfgSources, staticCredsLoader)
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), cfgSources...)
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
