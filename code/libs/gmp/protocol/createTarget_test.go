package protocol

// import (
// 	"context"
// 	"fmt"
// 	"testing"
// )

// func TestCreateTargets(t *testing.T) {
// 	api, err := NewOpenvasAPI(OpenvasConfig{
// 		Username: "admin",
// 		Password: "admin",
// 		CLIPath:  "/usr/local/bin/gvm-cli",
// 		Host:     "127.0.0.1",
// 		Port:     9390,
// 		Timeout:  600,
// 	})

// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	targets, err := api.CreateTarget(context.Background(), "test_name", "127.0.0.1", "test_comment")

// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	t.Log("create Targets test passed")
// 	fmt.Println(targets)
// }
