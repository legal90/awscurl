# awscurl

`awscurl` is a CLI tool allowing to send HTTP requests to AWS API. It automatically signs your requests with
[AWS Signature Version 4](https://docs.aws.amazon.com/general/latest/gr/signing_aws_api_requests.html),
so AWS can identify and authorize your request.

This implementation of [awscurl](https://github.com/okigan/awscurl) tool is written in Go.
It supports all AWS authentication methods available in AWS SDK for Go [v2](https://docs.aws.amazon.com/sdk-for-go/v2/api/), including:
- [AssumeRole profiles](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-role.html)
- IAM roles for [Amazon EC2 Instances](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html)
or [ECS Tasks](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html)

## Installation

### Cross-platform Binary (Recommended)

Download appropriate version for your platform from [awscurl Releases](https://github.com/legal90/awscurl/releases).
Once downloaded and unpacked, the binary can be run from anywhere. For convinience, you can place the binary to `/usr/local/bin/`

### Docker
awscurl container images are also released on DockerHub as [`legal90/docker`](https://hub.docker.com/repository/docker/legal90/awscurl)

```
$ docker run --rm \
    --env AWS_ACCESS_KEY_ID \
    --env AWS_SECRET_ACCESS_KEY \
    legal90/awscurl --help   # see Usage section for other available arguments
```
_Note:_ This command assumes that you have `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` exported in your environment.
You can find more details about AWS authentication in the _Usage_ section below.

### Build from Source

#### Prerequisites

- [Git](https://git-scm.com/)
- [Go (at least Go v1.13)](https://golang.org/dl/)

#### Fetch from GitHub

`awscurl` uses the Go Modules support built into Go 1.11. The easiest way to get started is to clone awscurl in a directory
outside of the $GOPATH, as in the following example:
```shell
mkdir $HOME/src
cd $HOME/src
git clone https://github.com/legal90/awscurl.git
cd awscurl
go build   # or `go install` to install it to your $GOPATH/bin
```

## Usage

```
$ awscurl --help

A simple CLI utility with cURL-like syntax allowing to send HTTP requests to AWS resources.
It automatically adds Siganture Version 4 to the request. More details:
https://docs.aws.amazon.com/general/latest/gr/signature-version-4.html

Usage:
  awscurl [URL] [flags]

Flags:
      --access-key string      AWS Access Key ID to use for authentication
  -d, --data string            Data payload to send within a POST request
  -H, --header stringArray     Extra HTTP header to include in the request. Example: "Content-Type: application/json". Could be used multiple times
      --profile string         AWS awsProfile to use for authentication
      --region string          AWS region to use for the request
  -X, --request string         Custom request method to use (default "GET")
      --secret-key string      AWS Secret Access Key to use for authentication
      --service string         The name of AWS Service, used for signing the request (default "execute-api")
      --session-token string   AWS Session Key to use for authentication
  -h, --help                   help for awscurl
  -v, --version                version for awscurl
```

### AWS Authentication
As you can see above, `awscurl` supports several argument, allowing you to choose the desired way to authenticate on AWS.
You can also use common environmental variables instead:

| CLI option          | Environment variable    |
|---------------------|-------------------------|
| `--access-key`      | `AWS_ACCESS_KEY_ID`     |
| `--secret-key`      | `AWS_SECRET_ACCESS_KEY` |
| `--session-token`   | `AWS_SESSION_TOKEN`     |
| `--profile`         | `AWS_PROFILE`           |
| `--region`          | `AWS_REGION`            |

By default, none of these variables are defined and AWS SDK for Go (used in `awscurl`)
will follow "the default provider chain". It looks for credentials in this order:

1. Environment variables (see table above).
2. Shared config and credentials file (`~/.aws/config`, `~/.aws/credentials`)
3. IAM role for Amazon EC2 or Tasks (if you run `awscurl` on EC2 Instance or ECS task)

### Examples

#### Call S3: List bucket content

This example also shows how to use a custom AWS profile "test".
Please note that it has to be configured in your `~/.aws/config` or `~/.aws/credentials`.
```shell
$ awscurl --service s3 \
    --profile "test" \
    "https://awscurl-sample-bucket.s3.amazonaws.com"
```

#### Call EC2:

In this example we also pass static AWS credentials using CLI arguments:
```shell
$ awscurl --service ec2 \
    --access-key <your-aws-access-key-id> \
    --secret-key <your-aws-secret-access-key> \
    "https://ec2.amazonaws.com?Action=DescribeRegions&Version=2013-10-15"
```

#### Call API Gateway:
```shell
$ awscurl --service execute-api \
    -X POST \
    -d '{"key": "value"}' \
    -H "Content-Type: application/json" \
    "https://<prefix>.execute-api.us-east-1.amazonaws.com/<resource>"
```
or reading data from the file:
```
$ awscurl --service execute-api \
    -X POST \
    -d @./path/to/file.json \
    -H "Content-Type: application/json" \
    "https://<prefix>.execute-api.us-east-1.amazonaws.com/<resource>"
```

## Related projects

- awscurl in Python: https://github.com/okigan/awscurl
- awscurl in Lisp: https://github.com/aw/picolisp-awscurl
- awscurl in Go (older implementation): https://github.com/allthings/awscurl

## License
[The MIT License](./LICENSE)

Copyright © 2020 Mikhail Zholobov <legal90@gmail.com>
