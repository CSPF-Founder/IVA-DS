package zap

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
)

type Ascan struct {
	c *Client
}

func NewAscan(c *Client) Ascan {
	return Ascan{c: c}
}

type AScanOpts struct {
	Recurse        string
	InScopeOnly    string
	ScanPolicyName string
	Method         string
	PostData       string
	ContextID      string
}

func (a Ascan) Scan(ctx context.Context, url string, opts AScanOpts) (string, error) {
	m := map[string]string{
		"url":            url,
		"recurse":        opts.Recurse,
		"inScopeOnly":    opts.InScopeOnly,
		"scanPolicyName": opts.ScanPolicyName,
		"method":         opts.Method,
		"postData":       opts.PostData,
		"contextId":      opts.ContextID,
	}

	res, err := a.c.Request(ctx, "ascan/action/scan/", m)
	if err != nil {
		return "", err
	}

	resJSON := struct {
		ScanID string `json:"scan"`
	}{}

	if err := json.Unmarshal(res, &resJSON); err != nil {
		return "", err
	}
	if resJSON.ScanID != "" {
		return resJSON.ScanID, nil
	}
	return "", errors.New("result not found")
}

func (a Ascan) Status(ctx context.Context, scanID string) (int, error) {
	m := map[string]string{
		"scanId": scanID,
	}
	res, err := a.c.Request(ctx, "ascan/view/status/", m)
	if err != nil {
		return 0, err
	}

	resJSON := struct {
		Status string `json:"status"`
	}{}

	if err := json.Unmarshal(res, &resJSON); err != nil {
		return 0, err
	}

	if resJSON.Status == "" {
		return 0, errors.New("scan status not found")
	}

	return strconv.Atoi(resJSON.Status)
}

func (a Ascan) Stop(ctx context.Context, scanID string) error {
	m := map[string]string{
		"scanId": scanID,
	}
	res, err := a.c.Request(ctx, "ascan/action/stop/", m)
	if err != nil {
		return err
	}

	resJSON := struct {
		Result string `json:"Result"`
	}{}

	if err := json.Unmarshal(res, &resJSON); err != nil {
		return errors.New("failed to parse response")
	}

	if resJSON.Result == "OK" {
		return nil
	}
	return errors.New("result not found")
}

func (a Ascan) RemoveScan(ctx context.Context, scanID string) error {
	m := map[string]string{
		"scanId": scanID,
	}
	res, err := a.c.Request(ctx, "ascan/action/removeScan/", m)
	if err != nil {
		return err
	}

	resJSON := struct {
		Result string `json:"Result"`
	}{}

	if err := json.Unmarshal(res, &resJSON); err != nil {
		return errors.New("failed to parse response")
	}

	if resJSON.Result == "OK" {
		return nil
	}
	return errors.New("result not found")
}
