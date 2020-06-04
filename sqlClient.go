package main

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	_ "github.com/denisenkom/go-mssqldb"
)

// DBConfig ...
type DBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// SQLClient ...
type SQLClient struct {
	*DBConfig
	err error
}

func createSQLClient(client secretsmanageriface.SecretsManagerAPI, dbSecretID string) (smClient *SQLClient) {
	smClient = &SQLClient{}

	if dbSecretID == "" {
		smClient.err = errors.New("Missing required 'SECRET_ID' parameter")
		return
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(dbSecretID),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}
	log.Print("Getting Secret")
	result, err := client.GetSecretValue(input)
	if err != nil {
		smClient.err = err
		return
	}

	var secretString string
	if result.SecretString != nil {
		secretString = *result.SecretString
	}
	if secretString == "" {
		smClient.err = errors.New("Unable to use secret")
		return
	}
	log.Print("Converting Secret to json")
	data := &DBConfig{}
	smClient.err = json.Unmarshal([]byte(secretString), data)
	smClient.DBConfig = data

	return
}

// Process ...
func (c *SQLClient) Process(event cfn.Event) (physicalResourceID string, jsonObject map[string]interface{}, err error) {
	if c.err != nil {
		err = c.err
		return
	}
	// eventJSON, _ := json.MarshalIndent(event, "", "  ")
	// log.Printf("EVENT: %s\n", string(eventJSON))

	// jsonObject = map[string]interface{}{}
	// physicalResourceID = event.PhysicalResourceID

	// switch event.RequestType {
	// case cfn.RequestCreate:
	// 	sqlClient, sqlerr := createSQLClient(event, smClient)
	// 	if sqlerr != nil {
	// 		log.Printf("ERROR: %s", sqlerr.Error())
	// 		err = sqlerr
	// 		return
	// 	}
	// 	physicalResourceID = sqlClient.Query
	// 	defer sqlClient.Close()

	// 	if err = sqlClient.run(); err != nil {
	// 		log.Printf("ERROR: %s", err.Error())
	// 		return
	// 	}
	// default:
	// 	log.Printf("Ignore: %s", event.RequestType)
	// }

	// return
	return
}

// func validate(event cfn.Event) (database string, sqlQuery string, dbSecretID string, err error) {
// 	var ok bool
// 	database, ok = event.ResourceProperties["Database"].(string)
// 	if !ok {
// 		err = errors.New("Missing 'Database' parameter")
// 		return
// 	}
// 	sqlQuery, ok = event.ResourceProperties["SqlQuery"].(string)
// 	if !ok {
// 		err = errors.New("Missing 'SqlQuery' parameter")
// 		return
// 	}
// 	dbSecretID, ok = os.Getenv("SECRET_ID")
// 	if !ok {
// 		err = errors.New("Missing 'SECRET_ID'")
// 		return
// 	}
// 	return
// }

// func createDBConfig(dbSecretID string, client secretsmanageriface.SecretsManagerAPI) (*DBConfig, error) {

// 	input := &secretsmanager.GetSecretValueInput{
// 		SecretId:     aws.String(dbSecretID),
// 		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
// 	}
// 	log.Print("Getting Secret")
// 	result, err := client.GetSecretValue(input)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var secretString string
// 	if result.SecretString != nil {
// 		secretString = *result.SecretString
// 	}
// 	if secretString == "" {
// 		return nil, errors.New("Unable to use secret")
// 	}
// 	log.Print("Converting Secret to json")
// 	data := &DBConfig{}
// 	err = json.Unmarshal([]byte(secretString), data)

// 	return data, err
// }

// func createSQLClient(event cfn.Event, client secretsmanageriface.SecretsManagerAPI) (*SQLClient, error) {

// 	dbName, sqlQuery, dbSecretID, err := validate(event)
// 	if err != nil {
// 		return nil, err
// 	}

// 	dbConfig, err := createDBConfig(dbSecretID, client)
// 	if err != nil {
// 		return nil, err
// 	}

// 	connString := fmt.Sprintf("Server=%s;Port=%d;Database=%s;User Id=%s;password=%s; Connection Timeout=%v", dbConfig.Host, dbConfig.Port, dbName, dbConfig.Username, dbConfig.Password, 5)
// 	log.Printf("ConnectionString: %s\n", connString)
// 	dbConn, err := sql.Open("sqlserver", connString)
// 	if err != nil {
// 		return nil, err
// 	}

// 	dbClient := &SQLClient{
// 		conn:  dbConn,
// 		Query: sqlQuery,
// 	}
// 	return dbClient, nil
// }

// // Close ...
// func (client *SQLClient) Close() {
// 	client.conn.Close()
// }

// func (client *SQLClient) run() error {
// 	log.Println("Begin Tx")
// 	tx, err := client.conn.Begin()
// 	if err != nil {
// 		return err
// 	}
// 	_, err = tx.Exec(client.Query)
// 	if err != nil {
// 		log.Println("Fail Tx")
// 		tx.Rollback()
// 		return err
// 	}
// 	err = tx.Commit()
// 	log.Println("End Tx")
// 	return err
// }
