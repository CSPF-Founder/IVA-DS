package protocol

import (
	"context"
	"encoding/xml"
	"errors"
)

type DeleteReportReq struct {
	XMLName  xml.Name `xml:"delete_report"`
	ReportID string   `xml:"report_id,attr" json:"report_id"`
}

type DeleteReportResp struct {
	*CommandResp
	XMLName xml.Name `xml:"delete_report_response"`
}

func (api *OpenvasAPI) DeleteReport(ctx context.Context, reportID string) error {
	req := DeleteReportReq{
		ReportID: reportID,
	}

	output, err := api.RunCmd(ctx, req)
	if err != nil {
		return err
	}

	resp := DeleteReportResp{}
	err = xml.Unmarshal(output, &resp)
	if err != nil {
		return err
	}

	if resp.Status == "200" || resp.Status == "202" {
		return nil
	} else if resp.Status == "404" {
		return errors.New("unable to find the report")
	}

	return errors.New("failed to delete the report")
}
