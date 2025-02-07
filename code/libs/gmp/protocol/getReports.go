package protocol

import (
	"context"
	"encoding/xml"
	"errors"
)

type GetReportsReq struct {
	XMLName          xml.Name `xml:"get_reports"`
	Filter           string   `xml:"filter,attr,omitempty"`
	ReportID         string   `xml:"report_id,attr" json:"report_id"`
	IgnorePagination bool     `xml:"ignore_pagination,attr" json:"ignore_pagination"`
	Details          string   `xml:"details,attr" json:"details"`
	FormatID         string   `xml:"format_id,attr" json:"format_id"`
}

type Report struct {
	ID      string `xml:"id,attr" csv:"id"`
	Content string `xml:",chardata" csv:"content"`
}

type GetReportsResp struct {
	*CommandResp
	XMLName xml.Name `xml:"get_reports_response"`
	Report  Report   `xml:"report"`
}

func (api *OpenvasAPI) GetReports(
	ctx context.Context,
	formatID string,
	reportID string,
	ignorePagination bool,
	enableDetails bool,
	filters string,
) (*Report, error) {
	details := "0"
	if enableDetails {
		details = "1"
	}

	req := GetReportsReq{
		FormatID:         formatID,
		ReportID:         reportID,
		IgnorePagination: ignorePagination,
		Details:          details,
		Filter:           filters,
	}
	output, err := api.RunCmd(ctx, req)
	if err != nil {
		return nil, err
	}

	resp := GetReportsResp{}
	err = xml.Unmarshal(output, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Status == "400" {
		return nil, errors.New(resp.StatusText)
	} else if resp.Status == "200" {
		content := &resp.Report
		return content, nil
	}

	return nil, errors.New(resp.StatusText)
}
