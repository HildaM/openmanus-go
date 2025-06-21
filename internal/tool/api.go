package tool

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
)

type APITool struct{}

func (a *APITool) Name() string {
	return "APITool"
}

// input: map[string]interface{}{"method": "GET"/"POST", "url": string, "body": string}
func (a *APITool) Run(ctx context.Context, input interface{}) (interface{}, error) {
	params, ok := input.(map[string]interface{})
	if !ok {
		return nil, ErrInvalidInputAPI
	}
	method, _ := params["method"].(string)
	url, _ := params["url"].(string)
	bodyStr, _ := params["body"].(string)

	var req *http.Request
	var err error
	if method == "POST" {
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer([]byte(bodyStr)))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return string(respBody), nil
}

var ErrInvalidInputAPI = &ToolError{"APITool: 输入必须为 map[method, url, body]"}
