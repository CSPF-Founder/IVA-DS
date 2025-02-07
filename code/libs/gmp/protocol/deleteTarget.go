package protocol

import (
	"context"
	"encoding/xml"
	"errors"
)

type DeleteTargetReq struct {
	XMLName  xml.Name `xml:"delete_target"`
	TargetID string   `xml:"target_id,attr" json:"target_id"`
}

type DeleteTargetResp struct {
	*CommandResp
	XMLName xml.Name `xml:"delete_target_response"`
}

func (api *OpenvasAPI) DeleteTarget(
	ctx context.Context,
	targetID string,
) error {
	// Command part
	req := DeleteTargetReq{
		TargetID: targetID,
	}

	output, err := api.RunCmd(ctx, req)
	if err != nil {
		return err
	}

	resp := DeleteTargetResp{}
	err = xml.Unmarshal(output, &resp)
	if err != nil {
		return err
	}

	if resp.Status == "200" || resp.Status == "202" {
		return nil
	} else if resp.Status == "400" {
		return errors.New("target is still being used")
	} else if resp.Status == "404" {
		return errors.New("unable to find the target")
	}

	return errors.New("unable to delete the target")
}
