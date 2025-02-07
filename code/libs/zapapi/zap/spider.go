package zap

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
)

type Spider struct {
	c *Client
}

func NewSpider(c *Client) Spider {
	return Spider{c: c}
}

func (s Spider) Status(ctx context.Context, scanID string) (int, error) {
	m := map[string]string{
		"scanId": scanID,
	}
	res, err := s.c.Request(ctx, "spider/view/status/", m)
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
		return 0, errors.New("spider status not found")
	}

	return strconv.Atoi(resJSON.Status)
}

func (s Spider) Scan(
	ctx context.Context,
	url string, maxchildren string,
	recurse string,
	contextname string,
	subtreeonly string,
) (string, error) {
	m := map[string]string{
		"url":         url,
		"maxChildren": maxchildren,
	}
	res, err := s.c.Request(ctx, "spider/action/scan/", m)

	if err != nil {
		return "", err
	}

	resJSON := struct {
		ScanID string `json:"scan"`
	}{}
	if err := json.Unmarshal(res, &resJSON); err != nil {
		return "", err
	}

	return resJSON.ScanID, nil
}

func (s Spider) StopAllScans(ctx context.Context) error {
	res, err := s.c.Request(ctx, "spider/action/stopAllScans/", nil)
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

func (s Spider) SetOptionParseRobotsTxt(ctx context.Context, boolean bool) error {
	m := map[string]string{
		"Boolean": strconv.FormatBool(boolean),
	}
	res, err := s.c.Request(ctx, "spider/action/setOptionParseRobotsTxt/", m)
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
	return errors.New("error in set option parse robots txt")
}

func (s Spider) SetOptionParseSitemapXml(ctx context.Context, boolean bool) error {
	m := map[string]string{
		"Boolean": strconv.FormatBool(boolean),
	}
	res, err := s.c.Request(ctx, "spider/action/setOptionParseSitemapXml/", m)
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

	return errors.New("error in set option parse sitemap xml")
}
