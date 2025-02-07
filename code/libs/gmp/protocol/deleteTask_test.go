package protocol

import (
	"context"
	"testing"
)

func TestDeleteTask(t *testing.T) {
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

	err = api.DeleteTask(context.Background(), "test_task_id")

	if err != nil {
		t.Fatal(err)
	}

	t.Log("DeleteTask test passed")
}
