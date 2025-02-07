package protocol

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"os/exec"
)

type OpenvasConfig struct {
	Username     string
	Password     string
	CLIPath      string
	ScanConfigID string
	PortListID   string
	Host         string
	Port         int
	Timeout      int
}

type OpenvasAPI struct {
	Config OpenvasConfig
}

func xmlToString(input any) (string, error) {
	data, err := xml.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("error marshalling xml: %v", err)
	}

	return string(data), nil
}

func (api OpenvasAPI) RunCmd(ctx context.Context, input any) ([]byte, error) {
	xmlStr, err := xmlToString(input)
	if err != nil {
		return nil, fmt.Errorf("error converting xml to string: %v", err)
	}

	cmd := exec.CommandContext(ctx,
		api.Config.CLIPath,
		"--timeout", fmt.Sprintf("%d", api.Config.Timeout),
		"--gmp-username", api.Config.Username,
		"--gmp-password", api.Config.Password,
		"tls",
		"--hostname", api.Config.Host,
		"--xml", xmlStr,
		"-r",
	)

	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	err = cmd.Run()
	if err != nil {
		fmt.Println(stdOut.String())
		if stdErr.String() != "" {
			return nil, fmt.Errorf("error running command: %v", stdErr.String())
		}
		return nil, fmt.Errorf("error running command: %v", err)
	}

	return stdOut.Bytes(), nil
}

func NewOpenvasAPI(config OpenvasConfig) (*OpenvasAPI, error) {
	if config.CLIPath == "" {
		return nil, fmt.Errorf("cli_path is required")
	}
	if config.Username == "" || config.Password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	if config.Host == "" {
		config.Host = "127.0.0.1"
	}

	if config.Port == 0 {
		config.Port = 9390
	}

	if config.Timeout == 0 {
		config.Timeout = 10
	}

	if config.ScanConfigID == "" {
		return nil, fmt.Errorf("ScanConfigId is required")
	}

	if config.PortListID == "" {
		return nil, fmt.Errorf("PortListId is required")
	}

	return &OpenvasAPI{Config: config}, nil
}
