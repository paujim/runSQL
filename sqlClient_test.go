package main

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aws/aws-lambda-go/cfn"
	"github.com/stretchr/testify/assert"
)

type mockSMClient struct {
	value string
	err   error
}

func (m *mockSMClient) GetSecretString(input string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.value, nil
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
		inputClient   CachedSecret
		inputEvent    cfn.Event
		expectedError string
	}{
		{&mockSMClient{}, cfn.Event{
			RequestType: cfn.RequestCreate,
			ResourceProperties: map[string]interface{}{
				"Database": "db",
				"SqlQuery": "sql",
			},
		}, "Missing required 'SecretId' parameter"},
		{&mockSMClient{}, cfn.Event{
			RequestType: cfn.RequestCreate,
			ResourceProperties: map[string]interface{}{
				"SqlQuery": "sql",
			},
		}, "Missing required 'Database' parameter"},
		{&mockSMClient{}, cfn.Event{
			RequestType: cfn.RequestCreate,
			ResourceProperties: map[string]interface{}{
				"Database": "db",
			},
		}, "Missing required 'SqlQuery' parameter"},
		{&mockSMClient{
			err: errors.New("Error from Secret"),
		}, cfn.Event{
			RequestType: cfn.RequestCreate,
			ResourceProperties: map[string]interface{}{
				"Database": "db",
				"SqlQuery": "sql",
				"SecretId": "secret",
			},
		}, "Error from Secret"},
		{&mockSMClient{}, cfn.Event{
			RequestType: cfn.RequestCreate,
			ResourceProperties: map[string]interface{}{
				"Database": "db",
				"SqlQuery": "sql",
				"SecretId": "secret",
			},
		}, "unexpected end of JSON input"},
		{&mockSMClient{
			value: "Invalid Json",
		}, cfn.Event{
			RequestType: cfn.RequestCreate,
			ResourceProperties: map[string]interface{}{
				"Database": "db",
				"SqlQuery": "sql",
				"SecretId": "secret",
			},
		}, "invalid character 'I' looking for beginning of value"},
	}

	for _, test := range tests {
		_, _, err := CreateLambdaHandler(test.inputClient, getDBConnection).Handle(test.inputEvent)
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
		value: "{\"host\": \"host\",\"username\": \"user\",\"password\": \"password\", \"port\":1344}",
	}, getDBConnection).Handle(cfn.Event{
		RequestType: cfn.RequestCreate,
		ResourceProperties: map[string]interface{}{
			"Database": "db",
			"SqlQuery": "UPDATE users FROM table",
			"SecretId": "secret",
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
		value: "{\"host\": \"host\",\"username\": \"user\",\"password\": \"password\", \"port\":1344}",
	}, getDBConnection).Handle(cfn.Event{
		RequestType: cfn.RequestCreate,
		ResourceProperties: map[string]interface{}{
			"Database": "db",
			"SqlQuery": "UPDATE users FROM table",
			"SecretId": "secret",
		}})
	assert.EqualError(t, err, "Wrong sql")
	assert.NoError(t, mock.ExpectationsWereMet())

}
