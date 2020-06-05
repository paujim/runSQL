package main

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/stretchr/testify/assert"
)

type mockSMClient struct {
	secretsmanageriface.SecretsManagerAPI
	value *string
	err   error
}

func (m *mockSMClient) GetSecretValue(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &secretsmanager.GetSecretValueOutput{SecretString: m.value}, nil
}

func TestLambdaHandler(t *testing.T) {

	getDBConnection := func(string) (*sql.DB, error) {
		return nil, nil
	}

	var tests = []struct {
		inputClient   secretsmanageriface.SecretsManagerAPI
		inputEvent    cfn.Event
		inputSecret   string
		expectedError string
	}{
		{&mockSMClient{}, cfn.Event{
			RequestType: cfn.RequestCreate,
			ResourceProperties: map[string]interface{}{
				"Database": "db",
				"SqlQuery": "sql",
			},
		}, "", "Missing required 'SECRET_ID' parameter"},
		{&mockSMClient{}, cfn.Event{
			RequestType: cfn.RequestCreate,
			ResourceProperties: map[string]interface{}{
				"SqlQuery": "sql",
			},
		}, "", "Missing required 'Database' parameter"},
		{&mockSMClient{}, cfn.Event{
			RequestType: cfn.RequestCreate,
			ResourceProperties: map[string]interface{}{
				"Database": "db",
			},
		}, "", "Missing required 'SqlQuery' parameter"},
		{&mockSMClient{
			err: errors.New("Error from Secret"),
		}, cfn.Event{
			RequestType: cfn.RequestCreate,
			ResourceProperties: map[string]interface{}{
				"Database": "db",
				"SqlQuery": "sql",
			},
		}, "Secret", "Error from Secret"},
		{&mockSMClient{
			value: nil,
		}, cfn.Event{
			RequestType: cfn.RequestCreate,
			ResourceProperties: map[string]interface{}{
				"Database": "db",
				"SqlQuery": "sql",
			},
		}, "Secret", "Unable to parse secret"},
		{&mockSMClient{
			value: aws.String(""),
		}, cfn.Event{
			RequestType: cfn.RequestCreate,
			ResourceProperties: map[string]interface{}{
				"Database": "db",
				"SqlQuery": "sql",
			},
		}, "Secret", "Unable to parse secret"},
		{&mockSMClient{
			value: aws.String("Invalid Json"),
		}, cfn.Event{
			RequestType: cfn.RequestCreate,
			ResourceProperties: map[string]interface{}{
				"Database": "db",
				"SqlQuery": "sql",
			},
		}, "Secret", "invalid character 'I' looking for beginning of value"},
	}

	for _, test := range tests {
		_, _, err := CreateLambdaHandler(test.inputClient, getDBConnection).Handle(test.inputSecret, test.inputEvent)
		assert.EqualError(t, err, test.expectedError, test)
	}

	// client := createSQLClient(&mockSMClient{value: aws.String("Not valid json")}, "Secret")
	// assert.Error(t, client.err, "Expected an error converting json")

	// client = createSQLClient(&mockSMClient{value: aws.String("Not valid json")}, "Secret")
	// assert.Error(t, client.err, "Expected an error converting json")

}
