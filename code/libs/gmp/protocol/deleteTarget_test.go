package protocol

import (
	"context"
	"testing"
)

func TestDeleteTarget(t *testing.T) {
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

	err = api.DeleteTarget(context.Background(), "test_target_id")

	if err != nil {
		t.Fatal(err)
	}

	t.Log("DeleteTarget test passed")
}
