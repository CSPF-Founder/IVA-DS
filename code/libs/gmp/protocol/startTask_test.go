package protocol

import (
	"context"
	"fmt"
	"testing"
)

func TestStartTask(t *testing.T) {
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

	task, err := api.StartTask(context.Background(), "test_task_id")

	if err != nil {
		t.Fatal(err)
	}

	t.Log("StartTask test passed")
	fmt.Println(task)
}
