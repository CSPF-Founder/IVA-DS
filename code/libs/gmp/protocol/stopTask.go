package protocol

import (
	"context"
	"encoding/xml"
	"errors"
)

type StopTaskReq struct {
	XMLName xml.Name `xml:"stop_task"`
	TaskID  string   `xml:"task_id,attr" json:"task_id"`
}

type StopTaskResp struct {
	*CommandResp
	XMLName xml.Name `xml:"stop_task_response"`
}

func (api *OpenvasAPI) StopTask(ctx context.Context, taskID string) error {
	req := StopTaskReq{
		TaskID: taskID,
	}

	output, err := api.RunCmd(ctx, req)
	if err != nil {
		return err
	}

	resp := StopTaskResp{}
	err = xml.Unmarshal(output, &resp)
	if err != nil {
		return err
	}

	if resp.Status == "200" || resp.Status == "202" {
		return nil
	}

	return errors.New("got non-200 status")
}
