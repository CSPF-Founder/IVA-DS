package protocol

import (
	"context"
	"encoding/xml"
	"errors"
)

type StartTaskReq struct {
	XMLName xml.Name `xml:"start_task"`
	TaskID  string   `xml:"task_id,attr" json:"task_id"`
}

type StartTaskResp struct {
	*CommandResp
	XMLName  xml.Name `xml:"start_task_response"`
	ReportID string   `xml:"report_id" json:"report_id"`
}

func (api *OpenvasAPI) StartTask(ctx context.Context, taskID string) (string, error) {
	req := StartTaskReq{
		TaskID: taskID,
	}
	output, err := api.RunCmd(ctx, req)
	if err != nil {
		return "", err
	}

	resp := StartTaskResp{}
	err = xml.Unmarshal(output, &resp)
	if err != nil {
		return "", err
	}

	if resp.Status == "400" {
		return "", errors.New(resp.StatusText)
	} else if resp.Status == "404" {
		return "", errors.New("failed to find the task")
	} else if resp.Status == "200" || resp.Status == "202" {
		if resp.ReportID == "" {
			return "", errors.New("got empty report id")
		}
		return resp.ReportID, nil
	}
	return "", errors.New(resp.StatusText)
}
