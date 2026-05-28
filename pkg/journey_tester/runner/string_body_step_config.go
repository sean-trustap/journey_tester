// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package runner

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"4d63.com/optional"
	journey_tester_context "github.com/trustap/journey_tester/pkg/journey_tester/context"
)

type StringBodyStepConfig struct {
	Req      StringBodyReq
	Resp     StrictResp
	Extracts Extracts
}

type StringBodyReq struct {
	URL       string
	Method    string
	Headers   *map[string]string
	Body      string
	URLQuery  *map[string]string
	Form      *map[string]string
	BasicAuth *BasicAuth
	File      *File
}

// FIXME Mostly duplicated from `newRequest`.
func (config StringBodyStepConfig) NewRequest(c *journey_tester_context.Context) (*http.Request, error) {
	reqConfig := config.Req

	body := strings.NewReader(reqConfig.Body)

	var req *http.Request
	if reqConfig.Form != nil {
		form := url.Values{}
		for k, v := range *reqConfig.Form {
			form.Set(k, v)
		}
		if len(form) != 0 {
			body = strings.NewReader(form.Encode())
		}
	}
	var contentType optional.Optional[string]

	req, err := http.NewRequest(reqConfig.Method, reqConfig.URL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if reqConfig.Headers != nil {
		for key, value := range *reqConfig.Headers {
			req.Header.Set(key, value)
		}
	}

	// TODO This may result in unexpected behaviour if the test attempts to
	// provide a custom `Request-ID` header value, as the provided
	// `Request-ID` value will be overridden. This should ideally be
	// documented explicitly.
	req.Header.Set("Request-ID", newRequestID())

	if reqConfig.URLQuery != nil {
		urlValues := url.Values{}
		for k, v := range *reqConfig.URLQuery {
			urlValues.Set(k, v)
		}
		if len(urlValues) != 0 {
			req.URL.RawQuery = urlValues.Encode()
		}
	}

	if reqConfig.BasicAuth != nil {
		req.SetBasicAuth(reqConfig.BasicAuth.Username, reqConfig.BasicAuth.Password)
	}

	if value, ok := contentType.Get(); ok {
		req.Header.Set("Content-Type", value)
	}

	return req, nil
}

func (config StringBodyStepConfig) ParseResponse(res *http.Response) (*ParsedResponse, error) {
	respConfig := config.Resp

	resp := Resp{
		Status:    respConfig.Status,
		Schema:    respConfig.Schema,
		Body:      respConfig.Body,
		RespFuncs: respConfig.RespFuncs,
	}

	return parseResponse(resp, res, respConfig.AllowExtraProperties)
}

func (config StringBodyStepConfig) Extract(c *journey_tester_context.Context, parsedResp *ParsedResponse) error {
	return extract(config.Extracts, c, parsedResp)
}
