package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/cfn"
	_ "github.com/denisenkom/go-mssqldb"
)

type CachedSecret interface {
	GetSecretString(string) (string, error)
}

// DBConfig ...
type DBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// LambdaHandler ...
type LambdaHandler struct {
	client          CachedSecret
	getDBConnection func(string) (*sql.DB, error)
}

// CreateLambdaHandler ..
func CreateLambdaHandler(client CachedSecret, getDBConnection func(string) (*sql.DB, error)) (smClient *LambdaHandler) {
	return &LambdaHandler{client: client, getDBConnection: getDBConnection}
}

// Handle ...
func (c *LambdaHandler) Handle(event cfn.Event) (physicalResourceID string, jsonObject map[string]interface{}, err error) {
	eventJSON, _ := json.MarshalIndent(event, "", "  ")
	log.Printf("EVENT: %s\n", string(eventJSON))

	jsonObject = map[string]interface{}{}
	physicalResourceID = event.PhysicalResourceID

	switch event.RequestType {
	case cfn.RequestCreate:
		var dbName, dbSecretID, connectionString, query string
		dbSecretID, dbName, query, err = c.validateParameters(event)
		if err != nil {
			return
		}
		connectionString, err = c.getConnectionString(dbSecretID, dbName)
		if err != nil {
			return
		}
		err = c.run(connectionString, query)
		if err != nil {
			return
		}
		physicalResourceID = c.getHash(dbName + query + dbSecretID)
	default:
		log.Printf("Ignore: %s", event.RequestType)
	}
	return
}

func (c *LambdaHandler) getHash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *LambdaHandler) validateParameters(event cfn.Event) (dbSecretID, dbName, sqlQuery string, err error) {
	var ok bool
	dbName, ok = event.ResourceProperties["Database"].(string)
	if !ok {

		err = errors.New("Missing required 'Database' parameter")
		return
	}
	sqlQuery, ok = event.ResourceProperties["SqlQuery"].(string)
	if !ok {
		err = errors.New("Missing required 'SqlQuery' parameter")
		return
	}
	dbSecretID, ok = event.ResourceProperties["SecretId"].(string)
	if !ok {
		err = errors.New("Missing required 'SecretId' parameter")
		return
	}
	return
}

func (c *LambdaHandler) getConnectionString(dbSecretID, dbName string) (connString string, err error) {
	log.Print("Getting Secret")
	secretString, err := c.client.GetSecretString(dbSecretID)
	if err != nil {
		return
	}

	data := &DBConfig{}
	err = json.Unmarshal([]byte(secretString), data)
	if err != nil {
		return
	}

	connString = fmt.Sprintf("Server=%s;Port=%d;Database=%s;User Id=%s;password=%s; Connection Timeout=%v", data.Host, data.Port, dbName, data.Username, data.Password, 5)
	log.Printf("ConnectionString: %s\n", connString) // FOR DEBBUGING, YOU MAY WANT TO REMOVE THIS LINE
	return connString, nil
}

func (c *LambdaHandler) run(connectionString, query string) error {

	dbConn, err := c.getDBConnection(connectionString)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	log.Println("Begin Tx")
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(query)
	if err != nil {
		log.Println("Fail Tx")
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	log.Println("End Tx")
	return err
}
