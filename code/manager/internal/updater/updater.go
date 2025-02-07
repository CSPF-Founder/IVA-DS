package updater

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/CSPF-Founder/iva/manager/utils"
)

const (
	defaultTimeout   = 20 * time.Minute
	defaultSleepTime = 5 * time.Minute
)

type UpdateType int

const (
	PANEL_FEED_UPDATE UpdateType = iota
	MANUAL_FEED_UPDATE
	NET_DOCKER_UPDATE
	WEB_DOCKER_UPDATE
)

func runCommand(ctx context.Context, cmd []string, timeout time.Duration, sleepTime time.Duration) error {
	command := exec.CommandContext(ctx, cmd[0], cmd[1:]...)

	var outBuf, errBuf bytes.Buffer
	command.Stdout = &outBuf
	command.Stderr = &errBuf

	// Run the command
	err := command.Run()

	// Capture output and error
	output := outBuf.String()
	errorOutput := errBuf.String()

	if err != nil {
		return err
	}

	if strings.Contains(output, "Re-creating") || strings.Contains(errorOutput, "Re-creating") {
		if err := utils.SleepContext(ctx, sleepTime*time.Second); err != nil {
			return fmt.Errorf("error sleeping: %w", err)
		}
	}
	return nil
}

func HandlePanelFeedUpdate(ctx context.Context, updaterCommand string) error {
	cmd := []string{
		updaterCommand,
		"-m",
		strconv.Itoa(int(MANUAL_FEED_UPDATE)),
	}
	err := runCommand(ctx, cmd, defaultTimeout, defaultSleepTime)
	if err != nil {
		return err
	}
	return nil
}

func UpdateNetworkFeed(ctx context.Context, updaterCommand string) error {
	cmd := []string{
		updaterCommand,
		"-m",
		strconv.Itoa(int(MANUAL_FEED_UPDATE)),
	}

	err := runCommand(ctx, cmd, defaultTimeout, defaultSleepTime)
	if err != nil {
		return err
	}
	return nil
}

func UpdateNetworkDocker(ctx context.Context, updaterCommand string) error {
	cmd := []string{
		updaterCommand,
		"-m",
		strconv.Itoa(int(NET_DOCKER_UPDATE)),
	}

	err := runCommand(ctx, cmd, defaultTimeout, defaultSleepTime)
	if err != nil {
		return err
	}
	return nil
}

func UpdateWebDocker(ctx context.Context, updaterCommand string) error {
	cmd := []string{
		updaterCommand,
		"-m",
		strconv.Itoa(int(WEB_DOCKER_UPDATE)),
	}

	err := runCommand(ctx, cmd, defaultTimeout, defaultSleepTime)
	if err != nil {
		return err
	}
	return nil
}
