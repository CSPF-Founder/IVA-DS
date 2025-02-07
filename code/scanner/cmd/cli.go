package main

import (
	"flag"
	"fmt"

	"github.com/CSPF-Founder/iva/scanner/utils"
)

type CLIInput struct {
	Module   string
	TargetID string
}

const (
	ModCheckNS = "check_ns"
	ModCheckWS = "check_ws"
	ModScan    = "scan"
)

func parseCLI() (CLIInput, error) {
	var cliInput CLIInput

	flag.StringVar(&cliInput.TargetID, "t", "", "target ID")
	flag.StringVar(&cliInput.Module, "m", "scan", "Module to check")

	flag.Parse()

	if cliInput.Module == ModScan {
		if !utils.IsValidObjectId(cliInput.TargetID) {
			return CLIInput{}, fmt.Errorf("invalid object id received")
		}
	}

	return cliInput, nil
}
