package protocol

import (
	"context"
	"fmt"
	"testing"
)

func TestGetVersion(t *testing.T) {
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

	ver, err := api.GetVersion(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	t.Log("GetVersion test passed")
	fmt.Println(ver)

}
