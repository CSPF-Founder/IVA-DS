package protocol

import (
	"context"
	"encoding/xml"
	"errors"
)

type GetTasksReq struct {
	XMLName          xml.Name  `xml:"get_tasks"`
	TaskID           string    `xml:"task_id,attr,omitempty"`      // ID of single task to get.
	Filter           string    `xml:"filter,omitempty"`            // Filter term to use to filter query.
	FiltID           string    `xml:"filt_id,omitempty"`           // ID of filter to use to filter query.
	Trash            bool      `xml:"trash,omitempty"`             // Whether to get the trashcan tasks instead.
	Details          bool      `xml:"details,omitempty"`           // Whether to include full task details.
	IgnorePagination bool      `xml:"ignore_pagination,omitempty"` // Whether to ignore info used to split the report into pages like the filter terms "first" and "rows"..
	SchedulesOnly    bool      `xml:"schedules_only,omitempty"`    // Whether to only include id, name and schedule details.
	UsageType        UsageType `xml:"usage_type,omitempty"`        // Optional usage type to limit the tasks to. Affects total count unlike filter.
}

type GetTaskListResp struct {
	*CommandResp
	XMLName        xml.Name               `xml:"get_tasks_response"`
	ApplyOverrides int                    `xml:"apply_overrides"`
	Tasks          []*GetTaskResp         `xml:"task" json:"task"`
	Filters        *Filters               `xml:"filters" json:"filters"`
	Sort           *Sort                  `xml:"sort" json:"sort"`
	TaskCount      *GetTasksRespTaskCount `xml:"task_count" json:"task_count"`
	// Progress       float32                `xml:"progress" json:"progress"`
}

type GetTaskResp struct {
	ID               string                     `xml:"id,attr" json:"id"`
	Owner            *Owner                     `xml:"owner" json:"owner"`                         // Owner of the task.
	Name             string                     `xml:"name" json:"name"`                           // The name of the task.
	Comment          string                     `xml:"comment" json:"comment"`                     // The comment on the task.
	CreationTime     string                     `xml:"creation_time" json:"creation_time"`         // Creation time of the task.
	ModificationTime string                     `xml:"modification_time" json:"modification_time"` // Last time the task was modified.
	Writable         bool                       `xml:"writable" json:"writable"`                   // Whether the task is writable.
	InUse            bool                       `xml:"in_use" json:"in_use"`                       // Whether this task is currently in use.
	Permissions      *Permissions               `xml:"permissions" json:"permissions"`             // Permissions that the current user has on the task.
	UserTags         *UserTags                  `xml:"user_tags" json:"user_tags"`                 // Info on tags attached to the task.
	Status           GetTasksRespStatus         `xml:"status" json:"status"`                       // The run status of the task.
	Progress         float32                    `xml:"progress" json:"progress"`                   // The percentage of the task that is complete.
	Alterable        bool                       `xml:"alterable" json:"alterable"`                 // Whether the task is an Alterable Task.
	UsageType        UsageType                  `xml:"usage_type" json:"usage_type"`               // The usage type of the task (scan or audit).
	Config           *GetTasksRespConfig        `xml:"config" json:"config"`                       // The scan configuration used by the task.
	Target           *GetTasksRespTarget        `xml:"target" json:"target"`                       // The hosts scanned by the task.
	HostsOrdering    string                     `xml:"hosts_ordering" json:"hosts_ordering"`       // The order hosts are scanned in.
	Scanner          *GetTasksRespScanner       `xml:"scanner" json:"scanner"`                     // The scanner used to scan the target.
	Alert            *GetTasksRespAlert         `xml:"alert" json:"alert"`                         // An alert that applies to the task.
	Observers        *GetTasksRespObservers     `xml:"observers" json:"observers"`                 // Users allowed to observe this task.
	Schedule         *GetTasksRespSchedule      `xml:"schedule" json:"schedule"`                   // When the task will run.
	SchedulePeriods  int                        `xml:"schedule_periods" json:"schedule_periods"`   // A limit to the number of times the task will be scheduled, or 0 for no limit.
	ReportCount      *GetTasksRespReportCount   `xml:"report_count" json:"report_count"`           // Number of reports.
	Trend            GetTasksRespTrend          `xml:"trend" json:"trend"`
	CurrentReport    *GetTasksRespCurrentReport `xml:"current_report" json:"current_report"`
	LastReport       *GetTasksRespLastReport    `xml:"last_report" json:"last_report"`
	AverageDuration  string                     `xml:"average_duration" json:"average_duration"` // Average scan duration in seconds.
	ResultCount      string                     `xml:"result_count" json:"result_count"`         // Result count for the entire task.
	Preferences      *GetTasksRespPreferences   `xml:"preferences" json:"preferences"`
}

type GetTasksRespStatus string

const (
	TaskStatusDeleteRequested GetTasksRespStatus = "Delete Requested"
	TaskStatusDone            GetTasksRespStatus = "Done"
	TaskStatusNew             GetTasksRespStatus = "New"
	TaskStatusRequested       GetTasksRespStatus = "Requested"
	TaskStatusRunning         GetTasksRespStatus = "Running"
	TaskStatusStopRequested   GetTasksRespStatus = "Stop Requested"
	TaskStatusStopped         GetTasksRespStatus = "Stopped"
	TaskStatusInterrupted     GetTasksRespStatus = "Interrupted"
)

type GetTasksRespTrend string

const (
	TaskTrendUp   GetTasksRespTrend = "up"
	TaskTrendDown GetTasksRespTrend = "down"
	TaskTrendMore GetTasksRespTrend = "more"
	TaskTrendLess GetTasksRespTrend = "less"
	TaskTrendSame GetTasksRespTrend = "same"
)

type GetTasksRespConfig struct {
	ID          string       `xml:"id,attr" json:"id"`
	Name        string       `xml:"name" json:"name"`               // The name of the config.
	Permissions *Permissions `xml:"permissions" json:"permissions"` // Permissions the user has on the config.
	Trash       bool         `xml:"trash" json:"trash"`             // Whether the config is in the trashcan.
}

type GetTasksRespTarget struct {
	ID          string       `xml:"id,attr" json:"id"`
	Name        string       `xml:"name" json:"name"`               // The hosts scanned by the task.
	Permissions *Permissions `xml:"permissions" json:"permissions"` // Permissions the user has on the target.
	Trash       bool         `xml:"trash" json:"trash"`             // Whether the target is in the trashcan.
}

type GetTasksRespScanner struct {
	ID          string       `xml:"id,attr" json:"id"`
	Name        string       `xml:"name" json:"name"`               // The name of the scanner.
	Permissions *Permissions `xml:"permissions" json:"permissions"` // Permissions the user has on the task.
	Type        int          `xml:"type" json:"type"`               // Type of the scanner.
}

type GetTasksRespAlert struct {
	ID          string       `xml:"id,attr" json:"id"`
	Name        string       `xml:"name" json:"name"`               // The name of the alert.
	Permissions *Permissions `xml:"permissions" json:"permissions"` // Permissions the user has on the alert.
	Trash       bool         `xml:"trash" json:"trash"`             // Whether the alert is in the trashcan.
}

type GetTasksRespObservers struct {
	Group []*GetTasksRespGroup `xml:"group" json:"group"` // Group allowed to observe this task.
	Role  []*GetTasksRespRole  `xml:"role" json:"role"`   // Role allowed to observe this task.
}

type GetTasksRespGroup struct {
	ID   string `xml:"id,attr" json:"id"`
	Name string `xml:"name" json:"name"` // The name of the group.
}

type GetTasksRespRole struct {
	ID   string `xml:"id,attr" json:"id"`
	Name string `xml:"name" json:"name"` // The name of the role.
}

type GetTasksRespSchedule struct {
	ID        string `xml:"id,attr" json:"id"`
	Name      string `xml:"name" json:"name"`           // The name of the schedule.
	Trash     bool   `xml:"trash" json:"trash"`         // Whether the schedule is in the trashcan.
	Icalendar string `xml:"icalendar" json:"icalendar"` // iCalendar text containing the time data..
	Timezone  string `xml:"timezone" json:"timezone"`   // The timezone the schedule will follow..
}

type GetTasksRespReportCount struct {
	Finished int `xml:"finished" json:"finished"` // Number of reports where the scan completed.
}

type GetTasksRespCurrentReport struct {
	Report *GetTasksRespReport `xml:"report" json:"report"`
}

type GetTasksRespReport struct {
	ID        string `xml:"id,attr" json:"id"`
	Timestamp string `xml:"timestamp" json:"timestamp"`
}

type GetTasksRespLastReport struct {
	Report *GetTasksRespLastReportDetail `xml:"report" json:"report"`
}

type GetTasksRespLastReportDetail struct {
	*GetTasksRespReport
	ScanEnd     string                   `xml:"scan_end" json:"scan_end"`
	ResultCount *GetTasksRespResultCount `xml:"result_count" json:"result_count"` // Result counts for this report.
	Severity    float32                  `xml:"severity" json:"severity"`         // Result count for the entire task.
}

type GetTasksRespResultCount struct {
	FalsePositive int `xml:"false_positive" json:"false_positive"`
	Log           int `xml:"log" json:"log"`
	Info          int `xml:"info" json:"info"`
	Warning       int `xml:"warning" json:"warning"`
	Hole          int `xml:"hole" json:"hole"`
}

type GetTasksRespPreferences struct {
	Preference []*GetTasksRespPreference `xml:"preference" json:"preference"`
}

type GetTasksRespPreference struct {
	Name        string `xml:"name" json:"name"`                 // Full name of preference, suitable for end users.
	ScannerName string `xml:"scanner_name" json:"scanner_name"` // Compact name of preference, from scanner.
	Value       string `xml:"value" json:"value"`
}

type GetTasksRespTasks struct {
	Start int `xml:"start,attr" json:"start"` // First task.
	Max   int `xml:"max,attr" json:"max"`     // Maximum number of tasks.

}

type GetTasksRespTaskCount struct {
	Filtered int `xml:"filtered" json:"filtered"` // Number of tasks after filtering.
	Page     int `xml:"page" json:"page"`         // Number of tasks on current page.
}

func (api *OpenvasAPI) GetTasks(ctx context.Context, filter string) ([]string, error) {
	req := GetTasksReq{
		Filter: filter,
	}
	output, err := api.RunCmd(ctx, req)
	if err != nil {
		return nil, err
	}

	resp := GetTaskListResp{}
	err = xml.Unmarshal(output, &resp)
	if err != nil {
		return nil, err
	}

	var taskIDs []string
	for _, task := range resp.Tasks {
		taskIDs = append(taskIDs, task.ID)
	}

	return taskIDs, nil
}

func (api *OpenvasAPI) GetProgress(ctx context.Context, taskID string) (int, error) {
	req := GetTasksReq{
		TaskID: taskID,
	}

	output, err := api.RunCmd(ctx, req)
	if err != nil {
		return 0, err
	}

	resp := GetTaskListResp{}
	err = xml.Unmarshal(output, &resp)
	if err != nil {
		return 0, err
	}

	if len(resp.Tasks) == 0 {
		return 0, errors.New("No tasks found")
	}

	task := resp.Tasks[0]

	err = checkStatusFailure(task.Status)
	if err != nil {
		return 0, err
	}

	return int(task.Progress), nil
}

// checkStatusNotFailed checks if the task status is not failed.
// it returns nil if the status matches one of the following:
// TaskStatusDone, TaskStatusNew, TaskStatusRequested, TaskStatusRunning
// otherwise it returns an error.
func checkStatusFailure(status GetTasksRespStatus) error {
	switch status {
	case TaskStatusDone, TaskStatusNew, TaskStatusRequested, TaskStatusRunning:
		return nil
	case TaskStatusDeleteRequested:
		return errors.New("Task delete requested")
	case TaskStatusStopRequested:
		return errors.New("Task stop requested")
	case TaskStatusStopped:
		return errors.New("Task stopped")
	case TaskStatusInterrupted:
		return errors.New("Task interrupted")
	default:
		return errors.New("Unknown task status")
	}
}
