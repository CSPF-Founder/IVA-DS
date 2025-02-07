package protocol

import (
	"context"
	"encoding/xml"
	"errors"
)

type DeleteCredential struct {
	XMLName      xml.Name `xml:"delete_lsc_credential"`
	CredentialID string   `xml:"credential_id" json:"credential_id"`
}

type DeleteCredentialResp struct {
	*CommandResp
	XMLName xml.Name `xml:"delete_credential_response"`
}

func (api *OpenvasAPI) DeleteCredential(ctx context.Context, credentialID string) error {
	req := &DeleteCredential{
		CredentialID: credentialID,
	}

	output, err := api.RunCmd(ctx, req)
	if err != nil {
		return err
	}

	resp := DeleteCredentialResp{}
	err = xml.Unmarshal(output, &resp)
	if err != nil {
		return err
	}

	if resp.Status == "200" || resp.Status == "202" {
		return nil
	} else if resp.Status == "400" {
		return errors.New("credential is still in use")
	} else if resp.Status == "404" {
		return errors.New("unable to find the credential entry")
	}

	return errors.New("unable to delete the credential")
}
