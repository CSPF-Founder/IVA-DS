package protocol

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
)

type CreateTargetReq struct {
	XMLName  xml.Name              `xml:"create_target"`
	Name     string                `xml:"name" json:"name"`
	Hosts    string                `xml:"hosts,omitempty" json:"hosts"`
	PortList *CreateTargetPortList `xml:"port_list,omitempty" json:"port_list"`
}

type CreateTargetRespTarget struct {
	XMLName        xml.Name                    `xml:"create_target"`
	Name           string                      `xml:"name" json:"name"`                 // A name for the target.
	Comment        string                      `xml:"comment,omitempty" json:"comment"` // A comment on the target.
	Copy           string                      `xml:"copy,omitempty" json:"copy"`
	AssetHosts     *CreateTargetAssetHosts     `xml:"asset_hosts,omitempty" json:"asset_hosts"`
	Hosts          string                      `xml:"hosts,omitempty" json:"hosts"`
	ExcludeHosts   string                      `xml:"exclude_hosts,omitempty" json:"exclude_hosts"`
	SshCredential  *CreateTargetSshCredential  `xml:"ssh_credential,omitempty" json:"ssh_credential"`
	SmbCredential  *CreateTargetSmbCredential  `xml:"smb_credential,omitempty" json:"smb_credential"`
	EsxiCredential *CreateTargetEsxiCredential `xml:"esxi_credential,omitempty" json:"esxi_credential"`
	// SNMP credentials to use on target.
	SnmpCredential *CreateTargetSnmpCredential `xml:"snmp_credential,omitempty" json:"snmp_credential"`
	// Which alive tests to use.
	AliveTests string `xml:"alive_tests,omitempty" json:"alive_tests"`
	// Whether to scan multiple IPs of the same host simultaneously.
	AllowSimultaneousIps bool `xml:"allow_simultaneous_ips,omitempty" json:"allow_simultaneous_ips"`
	// Whether to scan only hosts that have names.
	ReverseLookupOnly bool `xml:"reverse_lookup_only,omitempty" json:"reverse_lookup_only"`
	// Whether to scan only one IP when multiple IPs have the same name.
	ReverseLookupUnify bool                  `xml:"reverse_lookup_unify,omitempty" json:"reverse_lookup_unify"`
	PortRange          string                `xml:"port_range,omitempty" json:"port_range"`
	PortList           *CreateTargetPortList `xml:"port_list,omitempty" json:"port_list"`
}

type CreateTargetAssetHosts struct {
	Filter string `xml:"filter,attr" json:"filter"`
}

type CreateTargetSshCredential struct {
	ID   string `xml:"id,attr" json:"id"`
	Port string `xml:"port" json:"port"`
}

type CreateTargetSmbCredential struct {
	ID string `xml:"id,attr" json:"id"`
}

type CreateTargetEsxiCredential struct {
	ID string `xml:"id,attr" json:"id"`
}

type CreateTargetSnmpCredential struct {
	ID string `xml:"id,attr" json:"id"`
}

type CreateTargetPortList struct {
	ID string `xml:"id,attr" json:"id"`
}

type CreateTargetResp struct {
	*CommandResp
	XMLName xml.Name `xml:"create_target_response"`
	ID      string   `xml:"id,attr" json:"id"`
}

func (api *OpenvasAPI) CreateTarget(
	ctx context.Context,
	name string,
	host string,
	portList *CreateTargetPortList,
) (string, error) {
	// Command part
	req := CreateTargetReq{
		Name:     name,
		Hosts:    host,
		PortList: portList,
	}

	output, err := api.RunCmd(ctx, req)
	if err != nil {
		return "", err
	}

	resp := CreateTargetResp{}
	err = xml.Unmarshal(output, &resp)
	if err != nil {
		return "", err
	}
	if resp.Status == "400" {
		return "", errors.New(resp.StatusText)
	} else if resp.Status == "200" || resp.Status == "201" {
		if resp.ID == "" {
			return "", fmt.Errorf("got empty target id")
		}
		return resp.ID, nil
	}
	return "", errors.New(resp.StatusText)
}
