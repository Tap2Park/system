package manager

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// SetupVariables sets environment variables based on a secret retrieved from secret manager client.
// It takes a context and a secretPath as input.
// It returns an error if any operation fails.
func SetupVariables(ctx context.Context, secretPath string) error {

	manager := struct {
		BQDATASET  string `json:"BQ_DATASET"`
		BQEXTERNAL string `json:"BQ_EXTERNAL"`
	}{}

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create secretmanager client: %v", err)
	}

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretPath,
	}

	result, err := client.AccessSecretVersion(ctx, req)

	if err != nil {
		return fmt.Errorf("failed to access secret version: %v", err)
	}

	err = json.Unmarshal(result.Payload.Data, &manager)
	if err != nil {
		return fmt.Errorf("failed to access secret version: %v", err)
	}

	err = os.Setenv("BQ_DATASET", manager.BQDATASET)
	if err != nil {
		return fmt.Errorf("failed to set variable BQ_DATASET: %v", err)
	}
	err = os.Setenv("BQ_EXTERNAL", manager.BQEXTERNAL)
	if err != nil {
		return fmt.Errorf("failed to set variable BQ_EXTERNAL: %v", err)
	}

	return nil

}
