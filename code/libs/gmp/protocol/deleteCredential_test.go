package protocol

import (
	"context"
	"testing"
)

func TestDeleteCredential(t *testing.T) {
	api, err := NewOpenvasAPI(OpenvasConfig{
		Username: "admin",
		Password: "admin",
		CLIPath:  "/usr/local/bin/gvm-cli",
		Host:     "127.0.0.1",
		Port:     9390,
		Timeout:  600,
	})

	if err != nil {
		t.Fatal(err)
	}

	err = api.DeleteCredential(context.Background(), "cred_id")

	if err != nil {
		t.Fatal(err)
	}

	t.Log("DeleteReport test passed")
}
