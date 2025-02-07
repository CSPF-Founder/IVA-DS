package zap

import (
	"context"
	"encoding/json"
	"errors"
)

type ImportURLs struct {
	c *Client
}

func NewImportURLs(c *Client) ImportURLs {
	return ImportURLs{c: c}
}

func (i ImportURLs) FromFile(ctx context.Context, filepath string) error {
	m := map[string]string{
		"filePath": filepath,
	}
	res, err := i.c.Request(ctx, "importurls/action/importurls/", m)

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
