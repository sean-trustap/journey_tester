// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package tests

import (
	journey_tester_http "github.com/trustap/journey_tester/pkg/http"
	"github.com/trustap/journey_tester/pkg/journey_tester"
	journey_tester_context "github.com/trustap/journey_tester/pkg/journey_tester/context"
	"github.com/trustap/journey_tester/pkg/journey_tester/runner"
)

var DataJourneys = &journey_tester.DataJourneys{
	Create: []*journey_tester.Journey{
		CreateTestUser(
			"buyer",
			"users.buyer.email",
			"users.buyer.password",
		),
	},
	Verify: []*journey_tester.Journey{},
	Load:   []*journey_tester.Journey{},
	Delete: []*journey_tester.Journey{},
}

func CreateTestUser(
	userDescr string,
	emailVarName string,
	passwordVarName string,
) *journey_tester.Journey {
	j := journey_tester.New("Create " + userDescr)

	j.AddStep("Create new user", CreateUser(emailVarName, passwordVarName))

	return j
}

func CreateUser(emailVarName, passwordVarName string) journey_tester.StepFunc {
	return func(c *journey_tester_context.Context) runner.Step {
		return runner.StrictStepConfig{
			Req: runner.Req{
				URL:    c.Conf.Get("api.host") + "/v1/users",
				Method: "POST",
				Headers: &map[string]string{
					"Content-Type": journey_tester_http.ContentTypeApplicationXWwwFormURLencoded,
				},
				Form: &map[string]string{
					// FIXME `c.Conf.Get()` has special
					// behaviour here; update it to be
					// clearer.
					"email":    c.Conf.Get(emailVarName),
					"password": c.Conf.Get(passwordVarName),
				},
			},
			Resp: runner.StrictResp{
				Status: 200,
				Schema: &struct{}{},
				Body: &map[string]any{
					"id": nil,
				},
			},
			Extracts: runner.Extracts{
				ObjectPaths: map[string][]string{
					"test_seller_id": {"id"},
				},
			},
		}
	}
}
