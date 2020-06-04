package main

import (
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

func TestCreateSQLClient(t *testing.T) {
	var tests = []struct {
		inputClient   secretsmanageriface.SecretsManagerAPI
		inputSecret   string
		expectedError string
	}{
		{&mockSMClient{}, "", "Missing required 'SECRET_ID' parameter"},
		{&mockSMClient{err: errors.New("some Error")}, "Secret", "some Error"},
		{&mockSMClient{value: nil}, "Secret", "Unable to use secret"},
		{&mockSMClient{value: aws.String("")}, "Secret", "Unable to use secret"},
	}

	for _, test := range tests {
		assert.EqualError(t, createSQLClient(test.inputClient, test.inputSecret).err, test.expectedError)
	}

	client := createSQLClient(&mockSMClient{value: aws.String("Not valid json")}, "Secret")
	assert.Error(t, client.err, "Expected an error converting json")

	client = createSQLClient(&mockSMClient{value: aws.String("Not valid json")}, "Secret")
	assert.Error(t, client.err, "Expected an error converting json")

}

func TestProcess(t *testing.T) {

	sqlClient := SQLClient{err: errors.New("Has Error")}
	_, _, err := sqlClient.Process(cfn.Event{})
	assert.EqualError(t, err, "Has Error")

}
