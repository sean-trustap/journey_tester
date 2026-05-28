// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package runner

import (
	"net/http"

	journey_tester_context "github.com/trustap/journey_tester/pkg/journey_tester/context"
)

type StrictStepConfig struct {
	Req      Req
	Resp     StrictResp
	Extracts Extracts
}

type StrictResp struct {
	Status int
	// `Schema` is a struct that contains the response data.
	Schema any
	// `Body` is the expected data, by field and value.
	Body      any
	RespFuncs []RespFunc
	// Setting `AllowExtraProperties` to `false` effectively turns off
	// strict response checking. It should only be used when checking
	// responses from external services like Stripe, since the payload
	// itself doesn't need to be tested, only certain fields.
	AllowExtraProperties bool
}

func (config StrictStepConfig) NewRequest(c *journey_tester_context.Context) (*http.Request, error) {
	return newRequest(config.Req)
}

func (config StrictStepConfig) ParseResponse(res *http.Response) (*ParsedResponse, error) {
	respConfig := config.Resp

	resp := Resp{
		Status:    respConfig.Status,
		Schema:    respConfig.Schema,
		Body:      respConfig.Body,
		RespFuncs: respConfig.RespFuncs,
	}

	return parseResponse(resp, res, respConfig.AllowExtraProperties)
}

func (config StrictStepConfig) Extract(c *journey_tester_context.Context, parsedResp *ParsedResponse) error {
	return extract(config.Extracts, c, parsedResp)
}
