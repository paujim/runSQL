package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
)

var smClient secretsmanageriface.SecretsManagerAPI

func init() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	smClient = secretsmanager.New(sess)
}

func handler(ctx context.Context, event cfn.Event) (physicalResourceID string, jsonObject map[string]interface{}, err error) {
	return createSQLClient(smClient, os.Getenv("SECRET_ID")).Process(event)
}

func main() {
	lambda.Start(cfn.LambdaWrap(handler))
}
