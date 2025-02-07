package protocol

import (
	"context"
	"encoding/xml"
)

type GetVersionReq struct {
	XMLName xml.Name `xml:"get_version"`
}

type GetVersionResp struct {
	XMLName xml.Name `xml:"get_version_response"`
	*CommandResp

	Version string `xml:"version"`
}

func (api *OpenvasAPI) GetVersion(ctx context.Context) (string, error) {
	req := GetVersionReq{}
	output, err := api.RunCmd(ctx, req)
	if err != nil {
		return "", err
	}

	resp := GetVersionResp{}
	err = xml.Unmarshal(output, &resp)
	if err != nil {
		return "", err
	}

	return resp.Version, nil
}
