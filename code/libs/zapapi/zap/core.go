package zap

import (
	"context"
	"encoding/json"
	"errors"
)

type Core struct {
	client *Client
}

func NewCore(c *Client) Core {
	return Core{client: c}
}

type ScanType int

const (
	ScanTypeActive ScanType = iota
	ScanTypePassive
)

type AccessURLEntry struct {
	Note           string   `json:"note"`
	RTT            string   `json:"rtt"`
	ResponseBody   string   `json:"responseBody"`
	CookieParams   string   `json:"cookieParams"`
	RequestBody    string   `json:"requestBody"`
	ResponseHeader string   `json:"responseHeader"`
	RequestHeader  string   `json:"requestHeader"`
	ID             string   `json:"id"`
	Type           string   `json:"type"`
	Timestamp      string   `json:"timestamp"`
	Tags           []string `json:"tags"`
}

type Alert struct {
	Risk        string `json:"risk"`
	Title       string `json:"alert"`
	Description string `json:"description"`
	Solution    string `json:"solution"`
	WascID      string `json:"wascid"`
	CweID       string `json:"cweid"`
	MessageID   string `json:"messageId"`
	URL         string `json:"url"`
	Param       string `json:"param"`
	Reference   string `json:"reference"`
	Evidence    string `json:"evidence"`
	Attack      string `json:"attack"`
	Other       string `json:"other"`
	Instances   string `json:"instances"`
	Payload     string `json:"payload"`
	// Custom fields
	ScanType ScanType `json:"-"`
}
type Alerts struct {
	Entries []Alert `json:"alerts"`
}

type OptsToGetAlerts struct {
	BaseURL string
	Start   string
	Count   string
	RiskID  string
}

// Gets the alerts raised by ZAP, optionally filtering by URL or riskId, and paginating with 'start' position and 'count' of alerts
func (c Core) GetAlerts(ctx context.Context, opts OptsToGetAlerts) ([]Alert, error) {
	m := map[string]string{
		"baseurl": opts.BaseURL,
		"start":   opts.Start,
		"count":   opts.Count,
		"riskId":  opts.RiskID,
	}

	res, err := c.client.Request(ctx, "core/view/alerts/", m)
	if err != nil {
		return nil, err
	}

	var alerts Alerts

	err = json.Unmarshal(res, &alerts)
	if err != nil {
		return nil, err
	}

	return alerts.Entries, nil

}

// Creates a new session, optionally overwriting existing files. If a relative path is specified it will be resolved against the "session" directory in ZAP "home" dir.
func (c Core) NewSession(ctx context.Context, name string, overwrite string) error {
	m := map[string]string{
		"name":      name,
		"overwrite": overwrite,
	}
	res, err := c.client.Request(ctx, "core/action/newSession/", m)

	if err != nil {
		return err
	}

	resJSON := struct {
		Result string `json:"Result"`
	}{}
	if err := json.Unmarshal(res, &resJSON); err != nil {
		return err
	}

	if resJSON.Result == "OK" {
		return nil
	}
	return errors.New("status not found")
}

// Convenient and simple action to access a URL, optionally following redirections. Returns the request sent and response received and followed redirections, if any. Other actions are available which offer more control on what is sent, like, 'sendRequest' or 'sendHarRequest'.
func (c Core) AccessURL(ctx context.Context, url string, followredirects string) error {
	m := map[string]string{
		"url":             url,
		"followRedirects": followredirects,
	}
	res, err := c.client.Request(ctx, "core/action/accessUrl/", m)
	if err != nil {
		return err
	}

	resJSON := struct {
		Entries []AccessURLEntry `json:"accessUrl"`
	}{}

	if err := json.Unmarshal(res, &resJSON); err != nil {
		return err
	}

	if len(resJSON.Entries) > 0 {
		return nil
	}

	return errors.New("unable to reach")
}

// Gets ZAP version
func (c Core) Version(ctx context.Context) (string, error) {
	res, err := c.client.Request(ctx, "core/view/version/", nil)
	if err != nil {
		return "", err
	}

	resJSON := struct {
		Version string `json:"version"`
	}{}

	if err := json.Unmarshal(res, &resJSON); err != nil {
		return "", errors.New("failed to parse response")
	}

	if resJSON.Version != "" {
		return resJSON.Version, nil
	}

	return "", errors.New("version not found")
}
