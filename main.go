package main

import (
	"context"
	"database/sql"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
	_ "github.com/denisenkom/go-mssqldb"
)

var (
	secretCache, _ = secretcache.New()
)

func init() {
	secretCache, _ = secretcache.New()
}

func handler(ctx context.Context, event cfn.Event) (physicalResourceID string, jsonObject map[string]interface{}, err error) {
	getDBConnection := func(connectionString string) (*sql.DB, error) {
		return sql.Open("sqlserver", connectionString)
	}
	return CreateLambdaHandler(secretCache, getDBConnection).Handle(event)
}

func main() {
	lambda.Start(cfn.LambdaWrap(handler))
}
