package data

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"os"
	"time"
)

type DatabaseConfiguration struct {
	Host      string `json:"host"`
	PrivateIP string `json:"private"`
	PublicIP  string `json:"public"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Database  string `json:"database"`
}

// GetCredentialsFromSecretManager ------------------------------------------------------------------
func GetCredentialsFromSecretManager(ctx context.Context, secretPath string) (DatabaseConfiguration, error) {

	dbc := DatabaseConfiguration{}
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return dbc, fmt.Errorf("failed to create secretmanager client: %v", err)
	}

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretPath,
	}

	// Call the API.
	result, err := client.AccessSecretVersion(ctx, req)

	if err != nil {
		return dbc, fmt.Errorf("failed to access secret version: %v", err)
	}

	err = json.Unmarshal(result.Payload.Data, &dbc)

	return dbc, err

}

// GetCredentialsFromSecretEnvironmentVariable --------------------------------------------------------------
func GetCredentialsFromSecretEnvironmentVariable() (DatabaseConfiguration, error) {

	dbc := DatabaseConfiguration{}
	if len(os.Getenv("SECRET_PATH")) == 0 {
		return dbc, fmt.Errorf("missing SECRET_PATH Environment Variable")
	}

	err := json.Unmarshal([]byte(os.Getenv("SECRET_PATH")), &dbc)
	if err != nil {
		return dbc, fmt.Errorf("error parsing SECRET_PATH: %w", err)
	}

	return dbc, nil

}

// Connect -------------------------------------------------------
func Connect(dbc *DatabaseConfiguration) (*sql.DB, error) {

	if len(os.Getenv("DEVELOPMENT")) > 0 {
		return ConnectionByPublicIP(dbc)
	}

	if len(dbc.Host) == 0 {
		return nil, fmt.Errorf("missing Host")
	}

	sqlProxy := "cloudsql"

	if len(os.Getenv("SQLPROXY")) > 0 {
		sqlProxy = os.Getenv("SQLPROXY")
	}

	var dbURI string
	dbURI = fmt.Sprintf("%s:%s@unix(/%s/%s)/%s?autocommit=true&parseTime=true&timeout=5s", dbc.Username, dbc.Password, sqlProxy, dbc.Host, dbc.Database)

	// dbPool is the pool of database connections.
	link, err := sql.Open("mysql", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %v", err)
	}

	_sqlConnectionConfig(link)

	return link, err
}

// ConnectionByPublicIP --------------------------------------------------------------------------------------------------------------
func ConnectionByPublicIP(dbc *DatabaseConfiguration) (*sql.DB, error) {

	if len(dbc.PublicIP) == 0 {
		return nil, fmt.Errorf("missing PublicIP")
	}

	link, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?autocommit=true&parseTime=true&timeout=5s", dbc.Username, dbc.Password, dbc.PublicIP, dbc.Database))
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %v", err)
	}

	_sqlConnectionConfig(link)

	return link, err
}

// ConnectionByPrivateIP --------------------------------------------------------------------------------------------------------------
func ConnectionByPrivateIP(dbc *DatabaseConfiguration) (*sql.DB, error) {

	if len(dbc.PrivateIP) == 0 {
		return nil, fmt.Errorf("missing PrivateIP")
	}

	link, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?autocommit=true&parseTime=true&timeout=5s", dbc.Username, dbc.Password, dbc.PrivateIP, dbc.Database))
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %v", err)
	}

	_sqlConnectionConfig(link)

	return link, err
}

// _sqlConnectionConfig ----------------------------------------------------------------------------------
func _sqlConnectionConfig(link *sql.DB) {
	_, _ = link.Exec("SET time_zone = 'Europe/London'")
	// source: https://www.alexedwards.net/blog/configuring-sqldb
	link.SetMaxOpenConns(5)
	link.SetConnMaxIdleTime(2)
	link.SetConnMaxLifetime(1 * time.Hour)
}
