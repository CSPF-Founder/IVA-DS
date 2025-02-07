package protocol

import (
	"context"
	"encoding/xml"
	"errors"
)

type DeleteTaskReq struct {
	XMLName xml.Name `xml:"delete_task"`
	TaskID  string   `xml:"task_id,attr" json:"task_id"`
}

type DeleteTaskResp struct {
	*CommandResp
	XMLName xml.Name `xml:"delete_task_response"`
}

func (api *OpenvasAPI) DeleteTask(
	ctx context.Context,
	taskID string,
) error {
	// Command part
	req := DeleteTaskReq{
		TaskID: taskID,
	}

	output, err := api.RunCmd(ctx, req)
	if err != nil {
		return err
	}

	resp := DeleteTaskResp{}
	err = xml.Unmarshal(output, &resp)
	if err != nil {
		return err
	}

	if resp.Status == "400" {
		return errors.New("unable to find the task")
	} else if resp.Status == "200" || resp.Status == "202" {
		return nil
	}

	return errors.New("unable to delete the task")
}
