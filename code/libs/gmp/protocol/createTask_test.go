package protocol

import (
	"context"
	"fmt"
	"testing"
)

func TestCreateTask(t *testing.T) {
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

	task, err := api.CreateTask(context.Background(), "test_name", "test_target_id", "test_config_id")

	if err != nil {
		t.Fatal(err)
	}

	t.Log("CreateTask test passed")
	fmt.Println(task)
}
