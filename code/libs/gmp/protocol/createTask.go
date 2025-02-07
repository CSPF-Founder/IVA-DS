package protocol

import (
	"context"
	"encoding/xml"
	"errors"
)

type CreateTaskReq struct {
	XMLName xml.Name          `xml:"create_task"`
	Name    string            `xml:"name" json:"name"`
	Config  *CreateTaskConfig `xml:"config" json:"config"`
	Target  *CreateTaskTarget `xml:"target" json:"target"`
}

type CreateTask struct {
	XMLName         xml.Name               `xml:"create_task"`
	Name            string                 `xml:"name" json:"name"`                                   // A name for the task.
	Comment         string                 `xml:"comment,omitempty" json:"comment"`                   // A comment on the task.
	Copy            string                 `xml:"copy,omitempty" json:"copy"`                         // The UUID of an existing task.
	Alterable       bool                   `xml:"alterable,omitempty" json:"alterable"`               // Whether the task is alterable.
	UsageType       UsageType              `xml:"usage_type" json:"usage_type"`                       // Usage type for the task (scan or audit), defaulting to scan.
	Config          *CreateTaskConfig      `xml:"config" json:"config"`                               // The scan configuration used by the task.
	Target          *CreateTaskTarget      `xml:"target" json:"target"`                               // The hosts scanned by the task.
	HostsOrdering   string                 `xml:"hosts_ordering,omitempty" json:"hosts_ordering"`     // The order hosts are scanned in.
	Scanner         *CreateTaskScanner     `xml:"scanner" json:"scanner"`                             // The scanner to use for scanning the target.
	Alert           *CreateTaskAlert       `xml:"alert,omitempty" json:"alert"`                       // An alert that applies to the task.
	Schedule        *CreateTaskSchedule    `xml:"schedule,omitempty" json:"schedule"`                 // When the task will run.
	SchedulePeriods int                    `xml:"schedule_periods,omitempty" json:"schedule_periods"` // A limit to the number of times the task will be scheduled, or 0 for no limit.
	Observers       string                 `xml:"observers,omitempty" json:"observers"`               // Users allowed to observe this task.
	Preferences     *CreateTaskPreferences `xml:"preferences,omitempty" json:"preferences"`
}

type CreateTaskConfig struct {
	ID string `xml:"id,attr" json:"id"`
}

type CreateTaskTarget struct {
	ID string `xml:"id,attr" json:"id"`
}

type CreateTaskScanner struct {
	ID string `xml:"id,attr" json:"id"`
}

type CreateTaskAlert struct {
	ID string `xml:"id,attr" json:"id"`
}

type CreateTaskSchedule struct {
	ID string `xml:"id,attr" json:"id"`
}

type CreateTaskPreferences struct {
	Preference []*CreateTaskPreference `xml:"preference" json:"preference"`
}

type CreateTaskPreference struct {
	ScannerName string `xml:"scanner_name" json:"scanner_name"` // Compact name of preference, from scanner.
	Value       string `xml:"value" json:"value"`
}

type CreateTaskResp struct {
	*CommandResp
	XMLName xml.Name `xml:"create_task_response"`
	ID      string   `xml:"id,attr" json:"id"`
}

func (api *OpenvasAPI) CreateTask(
	ctx context.Context,
	name string,
	target_id string,
	config_id string,
) (string, error) {

	target := CreateTaskTarget{
		ID: target_id,
	}

	config := CreateTaskConfig{
		ID: config_id,
	}

	req := CreateTaskReq{
		Name:   name,
		Target: &target,
		Config: &config,
	}
	output, err := api.RunCmd(ctx, req)
	if err != nil {
		return "", err
	}

	resp := CreateTaskResp{}
	err = xml.Unmarshal(output, &resp)
	if err != nil {
		return "", err
	}

	if resp.Status == "400" {
		return "", errors.New(resp.StatusText)
	} else if resp.Status == "200" || resp.Status == "201" {
		if resp.ID == "" {
			return "", errors.New("got empty task id")
		}
		return resp.ID, nil
	}

	return "", errors.New(resp.StatusText)
}
