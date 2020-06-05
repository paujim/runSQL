package main

import (
	"context"
	"database/sql"
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

	getDBConnection := func(connectionString string) (*sql.DB, error) {
		return sql.Open("sqlserver", connectionString)
	}
	return CreateLambdaHandler(smClient, getDBConnection).Handle(os.Getenv("SECRET_ID"), event)
}

func main() {
	lambda.Start(cfn.LambdaWrap(handler))
}
