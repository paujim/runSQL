package main

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		return db, nil
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

}

func TestHandlerWithSQLSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	// Closes the database and prevents new queries from starting.
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	getDBConnection := func(string) (*sql.DB, error) {
		return db, nil
	}

	_, _, err = CreateLambdaHandler(&mockSMClient{
		value: aws.String("{\"host\": \"host\",\"username\": \"user\",\"password\": \"password\", \"port\":1344}"),
	}, getDBConnection).Handle("SecretId", cfn.Event{
		RequestType: cfn.RequestCreate,
		ResourceProperties: map[string]interface{}{
			"Database": "db",
			"SqlQuery": "UPDATE users FROM table",
		}})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

}

func TestHandlerWithSQLError(t *testing.T) {
	// Setup sql mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	// Closes the database and prevents new queries from starting.
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users").WillReturnError(errors.New("Wrong sql"))
	mock.ExpectRollback()

	getDBConnection := func(string) (*sql.DB, error) {
		return db, nil
	}

	_, _, err = CreateLambdaHandler(&mockSMClient{
		value: aws.String("{\"host\": \"host\",\"username\": \"user\",\"password\": \"password\", \"port\":1344}"),
	}, getDBConnection).Handle("SecretId", cfn.Event{
		RequestType: cfn.RequestCreate,
		ResourceProperties: map[string]interface{}{
			"Database": "db",
			"SqlQuery": "UPDATE users FROM table",
		}})
	assert.EqualError(t, err, "Wrong sql")
	assert.NoError(t, mock.ExpectationsWereMet())

}
